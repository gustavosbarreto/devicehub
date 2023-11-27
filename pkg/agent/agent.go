// Package agent provides packages and functions to create a new ShellHub Agent instance.
//
// The ShellHub Agent is a lightweight software component that runs the device and provide communication between the
// device and ShellHub's server. Its main role is to provide a reserve SSH server always connected to the ShellHub
// server, allowing SSH connections to be established to the device even when it is behind a firewall or NAT.
//
// This package provides a simple API to create a new agent instance and start the communication with the server. The
// agent will automatically connect to the server and start listening for incoming connections. Once connected, the
// agent will also automatically reconnect to the server if the connection is lost.
//
// The update process isn't handled by this package. This feature is provided by its main implementation in
// [ShellHub Agent]. Check the [ShellHub Agent] documentation for more information.
//
// # Example:
//
// Creates the agent configuration with the minimum required fields:
//
//	func main() {
//	    cfg := Config{
//	        ServerAddress: "http://localhost:80",
//	        TenantID:      "00000000-0000-4000-0000-000000000000",
//	        PrivateKey:    "/tmp/shellhub.key",
//	    }
//
//	    ctx := context.Background()
//	    ag, err := NewAgentWithConfig(&cfg)
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    if err := ag.Initialize(); err != nil {
//	        panic(err)
//	    }
//
//	    ag.Listen(ctx)
//	}
//
// [ShellHub Agent]: https://github.com/shellhub-io/shellhub/tree/master/agent
package agent

import (
	"context"
	"crypto/rsa"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/pkg/agent/server"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	log "github.com/sirupsen/logrus"
)

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option.
//
//	go build -ldflags "-X main.AgentVersion=1.2.3"
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

// AgentPlatform stores what platform the agent is running on. This is injected in build time in the [ShellHub Agent]
// implementation.
//
// [ShellHub Agent]: https://github.com/shellhub-io/shellhub/tree/master/agent
var AgentPlatform string

// Config provides the configuration for the agent service.
type Config struct {
	// Set the ShellHub Cloud server address the agent will use to connect.
	// This is required.
	ServerAddress string `env:"SERVER_ADDRESS,required"`

	// Specify the path to the device private key.
	// If not provided, the agent will generate a new one.
	// This is required.
	PrivateKey string `env:"PRIVATE_KEY,required"`

	// Sets the account tenant id used during communication to associate the
	// device to a specific tenant.
	// This is required.
	TenantID string `env:"TENANT_ID,required"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval int `env:"KEEPALIVE_INTERVAL,default=30"`

	// Set the device preferred hostname. This provides a hint to the server to
	// use this as hostname if it is available.
	PreferredHostname string `env:"PREFERRED_HOSTNAME"`

	// Set the device preferred identity. This provides a hint to the server to
	// use this identity if it is available.
	PreferredIdentity string `env:"PREFERRED_IDENTITY,default="`

	// Set password for single-user mode (without root privileges). If not provided,
	// multi-user mode (with root privileges) is enabled by default.
	// NOTE: The password hash could be generated by ```openssl passwd```.
	SingleUserPassword string `env:"SIMPLE_USER_PASSWORD"`
}

type Agent struct {
	config        *Config
	pubKey        *rsa.PublicKey
	Identity      *models.DeviceIdentity
	Info          *models.DeviceInfo
	authData      *models.DeviceAuthResponse
	cli           client.Client
	serverInfo    *models.Info
	serverAddress *url.URL
	sessions      []string
	server        *server.Server
	tunnel        *tunnel.Tunnel
	mux           sync.RWMutex
	listening     chan bool
	closed        bool
	mode          Mode
}

// NewAgent creates a new agent instance.
//
// address is the ShellHub Server address the agent will use to connect, tenantID is the namespace where the device
// will be registered and privateKey is the path to the device private key. If privateKey is empty, a new key will be
// generated.
//
// To add a full customisation configuration, use [NewAgentWithConfig] instead.
func NewAgent(address string, tenantID string, privateKey string, mode Mode) (*Agent, error) {
	return NewAgentWithConfig(&Config{
		ServerAddress: address,
		TenantID:      tenantID,
		PrivateKey:    privateKey,
	}, mode)
}

// NewAgentWithConfig creates a new agent instance with a custom configuration.
//
// Check [Config] for more information.
func NewAgentWithConfig(config *Config, mode Mode) (*Agent, error) {
	if config.ServerAddress == "" {
		return nil, errors.New("address is empty")
	}

	serverAddress, err := url.Parse(config.ServerAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse address")
	}

	cli, err := client.NewClient(config.ServerAddress)
	if err != nil {
		return nil, err
	}

	if config.TenantID == "" {
		return nil, errors.New("tenantID is empty")
	}

	if config.PrivateKey == "" {
		return nil, errors.New("privateKey is empty")
	}

	if mode == nil {
		return nil, errors.New("mode cannot be nil")
	}

	a := &Agent{
		config:        config,
		serverAddress: serverAddress,
		cli:           cli,
		listening:     make(chan bool),
		mode:          mode,
	}

	return a, nil
}

// Initialize initializes agent, generating device identity, loading device information, generating private key,
// reading public key, probing server information and authorizing device on ShellHub server.
//
// When any of the steps fails, the agent will return an error, and the agent will not be able to start.
func (a *Agent) Initialize() error {
	if err := a.generateDeviceIdentity(); err != nil {
		return errors.Wrap(err, "failed to generate device identity")
	}

	if err := a.loadDeviceInfo(); err != nil {
		return errors.Wrap(err, "failed to load device info")
	}

	if err := a.generatePrivateKey(); err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}

	if err := a.readPublicKey(); err != nil {
		return errors.Wrap(err, "failed to read public key")
	}

	if err := a.probeServerInfo(); err != nil {
		return errors.Wrap(err, "failed to probe server info")
	}

	if err := a.authorize(); err != nil {
		return errors.Wrap(err, "failed to authorize device")
	}

	a.mux.Lock()
	a.closed = false
	a.mux.Unlock()

	return nil
}

// generatePrivateKey generates a new private key if it doesn't exist on the filesystem.
func (a *Agent) generatePrivateKey() error {
	if _, err := os.Stat(a.config.PrivateKey); os.IsNotExist(err) {
		if err := keygen.GeneratePrivateKey(a.config.PrivateKey); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) readPublicKey() error {
	key, err := keygen.ReadPublicKey(a.config.PrivateKey)
	a.pubKey = key

	return err
}

// generateDeviceIdentity generates device identity.
//
// When preferred identity on Agent is set, it will be used instead of the network interface MAC address, what is the
// default value for this property.
func (a *Agent) generateDeviceIdentity() error {
	if id := a.config.PreferredIdentity; id != "" {
		a.Identity = &models.DeviceIdentity{
			MAC: id,
		}

		return nil
	}

	// get identity from network interface.
	iface, err := sysinfo.PrimaryInterface()
	if err != nil {
		return err
	}

	a.Identity = &models.DeviceIdentity{
		MAC: iface.HardwareAddr.String(),
	}

	return nil
}

// loadDeviceInfo load some device informations like OS name, version, arch and platform.
func (a *Agent) loadDeviceInfo() error {
	info, err := a.mode.GetInfo()
	if err != nil {
		return err
	}

	a.Info = &models.DeviceInfo{
		ID:         info.ID,
		PrettyName: info.Name,
		Version:    AgentVersion,
		Platform:   AgentPlatform,
		Arch:       runtime.GOARCH,
	}

	return nil
}

// probeServerInfo probe server information.
func (a *Agent) probeServerInfo() error {
	info, err := a.cli.GetInfo(AgentVersion)
	a.serverInfo = info

	return err
}

// authorize send auth request to the server.
func (a *Agent) authorize() error {
	data, err := a.cli.AuthDevice(&models.DeviceAuthRequest{
		Info: a.Info,
		DeviceAuth: &models.DeviceAuth{
			Hostname:  a.config.PreferredHostname,
			Identity:  a.Identity,
			TenantID:  a.config.TenantID,
			PublicKey: string(keygen.EncodePublicKeyToPem(a.pubKey)),
		},
	})

	a.authData = data

	return err
}

func (a *Agent) NewReverseListener(ctx context.Context) (*revdial.Listener, error) {
	return a.cli.NewReverseListener(ctx, a.authData.Token)
}

func (a *Agent) Close() error {
	a.mux.Lock()
	a.closed = true
	a.mux.Unlock()

	return a.tunnel.Close()
}

func connHandler(serv *server.Server) func(c echo.Context) error {
	return func(c echo.Context) error {
		hj, ok := c.Response().Writer.(http.Hijacker)
		if !ok {
			return c.String(http.StatusInternalServerError, "webserver doesn't support hijacking")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to hijack connection")
		}

		id := c.Param("id")
		httpConn := c.Request().Context().Value("http-conn").(net.Conn)
		serv.Sessions.Store(id, httpConn)
		serv.HandleConn(httpConn)

		conn.Close()

		return nil
	}
}

func httpHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		replyError := func(err error, msg string, code int) error {
			log.WithError(err).WithFields(log.Fields{
				"remote":    c.Request().RemoteAddr,
				"namespace": c.Request().Header.Get("X-Namespace"),
				"path":      c.Request().Header.Get("X-Path"),
				"version":   AgentVersion,
			}).Error(msg)

			return c.String(code, msg)
		}

		in, err := net.Dial("tcp", ":80")
		if err != nil {
			return replyError(err, "failed to connect to HTTP server on device", http.StatusInternalServerError)
		}

		defer in.Close()

		url, err := url.Parse(c.Request().Header.Get("X-Path"))
		if err != nil {
			return replyError(err, "failed to parse URL", http.StatusInternalServerError)
		}

		c.Request().URL.Scheme = "http"
		c.Request().URL = url

		if err := c.Request().Write(in); err != nil {
			return replyError(err, "failed to write request to the server on device", http.StatusInternalServerError)
		}

		out, _, err := c.Response().Hijack()
		if err != nil {
			return replyError(err, "failed to hijack connection", http.StatusInternalServerError)
		}

		defer out.Close() // nolint:errcheck

		if _, err := io.Copy(out, in); err != nil {
			return replyError(err, "failed to copy response from device service to client", http.StatusInternalServerError)
		}

		return nil
	}
}

func closeHandler(a *Agent, serv *server.Server) func(c echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")
		serv.CloseSession(id)

		log.WithFields(
			log.Fields{
				"id":             id,
				"version":        AgentVersion,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")

		return nil
	}
}

// Listen creates a new SSH server, through a reverse connection between the Agent and the ShellHub server.
func (a *Agent) Listen(ctx context.Context) error {
	a.mode.Serve(a)

	a.tunnel = tunnel.NewBuilder().
		WithConnHandler(connHandler(a.server)).
		WithCloseHandler(closeHandler(a, a.server)).
		WithHTTPHandler(httpHandler()).
		Build()

	done := make(chan bool)
	go func() {
		for {
			a.mux.RLock()
			if a.closed {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
				}).Info("Stopped listening for connections")

				done <- true

				a.mux.RUnlock()

				return
			}
			a.mux.RUnlock()

			namespace := a.authData.Namespace
			tenantName := a.authData.Name
			sshEndpoint := a.serverInfo.Endpoints.SSH

			sshid := strings.NewReplacer(
				"{namespace}", namespace,
				"{tenantName}", tenantName,
				"{sshEndpoint}", strings.Split(sshEndpoint, ":")[0],
			).Replace("{namespace}.{tenantName}@{sshEndpoint}")

			listener, err := a.NewReverseListener(ctx)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"ssh_server":     sshEndpoint,
					"sshid":          sshid,
				}).Error("Failed to connect to server through reverse tunnel. Retry in 10 seconds")
				time.Sleep(time.Second * 10)

				continue
			}

			log.WithFields(log.Fields{
				"namespace":      namespace,
				"hostname":       tenantName,
				"server_address": a.config.ServerAddress,
				"ssh_server":     sshEndpoint,
				"sshid":          sshid,
			}).Info("Server connection established")

			a.listening <- true

			if err := a.tunnel.Listen(listener); err != nil {
				// NOTICE: Tunnel'll only realize that it lost its connection to the ShellHub SSH when the next
				// "keep-alive" connection fails. As a result, it will take this interval to reconnect to its server.
				//
				// It can be observed in the logs, that prints something like:
				//  0000/00/00 00:00:00 revdial.Listener: error writing message to server: write tcp [::1]:00000->[::1]:80: write: broken pipe
				log.WithError(err).WithFields(log.Fields{
					"namespace":      namespace,
					"hostname":       tenantName,
					"server_address": a.config.ServerAddress,
					"ssh_server":     sshEndpoint,
					"sshid":          sshid,
				}).Error("Tunnel listener closed")

				listener.Close() // nolint:errcheck
				a.listening <- false

				continue
			}

			log.WithError(err).WithFields(log.Fields{
				"namespace":      namespace,
				"hostname":       tenantName,
				"server_address": a.config.ServerAddress,
				"ssh_server":     sshEndpoint,
				"sshid":          sshid,
			}).Info("Tunnel listener closed")

			listener.Close() // nolint:errcheck
			a.listening <- false
		}
	}()

	select {
	case <-ctx.Done():
		if err := a.Close(); err != nil {
			return err
		}

		return nil
	case <-done:
		return nil
	}
}

// Ping sends an authtorization request to the server every ticker interval.
//
// If the durantion is 0, the default value set to it will be the 10 minutes.
//
// Ping will only sends its requests to the server if the agent is listening for connections. If the agent is not
// listening, the ping will be stopped.
func (a *Agent) Ping(ctx context.Context, durantion time.Duration) error {
	if durantion == 0 {
		durantion = 10 * time.Minute
	}

	ticker := time.NewTicker(durantion)
	<-a.listening // NOTE: wait for the first connection to start to ping the server.

	for {
		a.mux.RLock()
		if a.closed {
			a.mux.RUnlock()

			return nil
		}
		a.mux.RUnlock()

		select {
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"version":        AgentVersion,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			}).Debug("stopped pinging server due to context cancellation")

			return nil
		case ok := <-a.listening:
			if ok {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"timestamp":      time.Now(),
				}).Info("Restarted pinging server")

				ticker.Reset(durantion)
			} else {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"timestamp":      time.Now(),
				}).Info("Stopped pinging server due listener status")

				ticker.Stop()
			}
		case <-ticker.C:
			var sessions []string
			a.server.Sessions.Range(func(k, _ interface{}) bool {
				sessions = append(sessions, k.(string))

				return true
			})

			a.sessions = sessions

			if err := a.authorize(); err != nil {
				a.server.SetDeviceName(a.authData.Name)
			}

			log.WithFields(log.Fields{
				"version":        AgentVersion,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
				"name":           a.authData.Name,
				"hostname":       a.config.PreferredHostname,
				"identity":       a.config.PreferredIdentity,
				"timestamp":      time.Now(),
			}).Info("Ping")
		}
	}
}

// CheckUpdate gets the ShellHub's server version.
func (a *Agent) CheckUpdate() (*semver.Version, error) {
	info, err := a.cli.GetInfo(AgentVersion)
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(info.Version)
}

// GetInfo gets the ShellHub's server information like version and endpoints, and updates the Agent's server's info.
func (a *Agent) GetInfo() (*models.Info, error) {
	if a.serverInfo != nil {
		return a.serverInfo, nil
	}

	info, err := a.cli.GetInfo(AgentVersion)
	if err != nil {
		return nil, err
	}

	a.serverInfo = info

	return info, nil
}
