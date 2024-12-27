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
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Masterminds/semver"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/pkg/agent/server"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
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
	ServerAddress             string `env:"SERVER_ADDRESS,required" validate:"required"`
	PrivateKey                string `env:"PRIVATE_KEY,required" validate:"required"`
	TenantID                  string `env:"TENANT_ID,required" validate:"required"`
	PreferredHostname         string `env:"PREFERRED_HOSTNAME"`
	PreferredIdentity         string `env:"PREFERRED_IDENTITY,default="`
	SingleUserPassword        string `env:"SINGLE_USER_PASSWORD,default=$SIMPLE_USER_PASSWORD"`
	SimpleUserPassword        string `env:"SIMPLE_USER_PASSWORD"`
	KeepAliveInterval         uint   `env:"KEEPALIVE_INTERVAL,overwrite,default=30"`
	MaxRetryConnectionTimeout int    `env:"MAX_RETRY_CONNECTION_TIMEOUT,default=60" validate:"min=10,max=120"`
}

func LoadConfigFromEnv() (*Config, map[string]interface{}, error) {
	// NOTE(r): When T, the generic parameter, is a structure with required tag, the fallback for an
	// "unprefixed" parameter is used.
	//
	// For example,
	//
	// For the structure below, the parser will parse successfully when the variables exist with or without the
	// prefixes since the "required" tag is set to true.
	//
	//  SHELLHUB_TENANT_ID=00000000-0000-4000-0000-000000000000 SERVER_ADDRESS=http://127.0.0.1
	//  PRIVATE_KEY=/tmp/shellhub sudo -E ./agent
	//
	//  struct {
	//    ServerAddress string `env:"SERVER_ADDRESS,required"`
	//    PrivateKey string `env:"PRIVATE_KEY,required"`
	//    TenantID string `env:"TENANT_ID,required`
	//  }
	//
	//  This behavior is driven by the [envconfig] package. Check it out for more information.
	//
	// [envconfig]: https://github.com/sethvargo/go-envconfig
	cfg, err := envs.ParseWithPrefix[Config]("SHELLHUB_")
	if err != nil {
		log.Error("failed to parse the configuration")

		return nil, nil, err
	}

	// TODO: test the envinromental variables validation on integration tests.
	if ok, fields, err := validator.New().StructWithFields(cfg); err != nil || !ok {
		log.WithFields(fields).Error("failed to validate the configuration loaded from envs")

		return nil, fields, err
	}

	return cfg, nil, nil
}

type Agent struct {
	cli        client.Client
	mode       Mode
	config     *Config
	pubKey     *rsa.PublicKey
	Identity   *models.DeviceIdentity
	Info       *models.DeviceInfo
	authData   *models.DeviceAuthResponse
	serverInfo *models.Info
	server     *server.Server
	tunnel     *tunnel.Tunnel
	listening  chan bool
	closed     atomic.Bool
}

// NewAgent creates a new agent instance, requiring the ShellHub server's address to connect to, the namespace's tenant
// where device own and the path to the private key on the file system.
//
// To create a new [Agent] instance with all configurations, you can use [NewAgentWithConfig].
func NewAgent(address string, tenantID string, privateKey string, mode Mode) (*Agent, error) {
	return NewAgentWithConfig(&Config{
		ServerAddress: address,
		TenantID:      tenantID,
		PrivateKey:    privateKey,
	}, mode)
}

var (
	ErrNewAgentWithConfigEmptyServerAddress   = errors.New("address is empty")
	ErrNewAgentWithConfigInvalidServerAddress = errors.New("address is invalid")
	ErrNewAgentWithConfigEmptyTenant          = errors.New("tenant is empty")
	ErrNewAgentWithConfigEmptyPrivateKey      = errors.New("private key is empty")
	ErrNewAgentWithConfigNilMode              = errors.New("agent's mode is nil")
)

// NewAgentWithConfig creates a new agent instance with all configurations.
//
// Check [Config] for more information.
func NewAgentWithConfig(config *Config, mode Mode) (*Agent, error) {
	if config.ServerAddress == "" {
		return nil, ErrNewAgentWithConfigEmptyServerAddress
	}

	if _, err := url.ParseRequestURI(config.ServerAddress); err != nil {
		return nil, ErrNewAgentWithConfigInvalidServerAddress
	}

	if config.TenantID == "" {
		return nil, ErrNewAgentWithConfigEmptyTenant
	}

	if config.PrivateKey == "" {
		return nil, ErrNewAgentWithConfigEmptyPrivateKey
	}

	if mode == nil {
		return nil, ErrNewAgentWithConfigNilMode
	}

	return &Agent{
		config: config,
		mode:   mode,
	}, nil
}

// Initialize initializes the ShellHub Agent, generating device identity, loading device information, generating private
// key, reading public key, probing server information and authorizing device on ShellHub server.
//
// When any of the steps fails, the agent will return an error, and the agent will not be able to start.
func (a *Agent) Initialize() error {
	var err error

	a.cli, err = client.NewClient(a.config.ServerAddress)
	if err != nil {
		return errors.Wrap(err, "failed to create the HTTP client")
	}

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

	a.closed.Store(false)

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

// generateDeviceIdentity generates a device identity.
//
// The default value for Agent Identity is a network interface MAC address, but if the `SHELLHUB_PREFERRED_IDENTITY` is
// defined and set on [Config] structure, the device identity is set to this value.
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

// probeServerInfo gets information about the ShellHub server.
func (a *Agent) probeServerInfo() error {
	info, err := a.cli.GetInfo(AgentVersion)
	a.serverInfo = info

	return err
}

// authorize send auth request to the server with device information in order to register it in the namespace.
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

func (a *Agent) isClosed() bool {
	return a.closed.Load()
}

// Close closes the ShellHub Agent's listening, stoping it from receive new connection requests.
func (a *Agent) Close() error {
	a.closed.Store(true)

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

// proxyHandler handlers proxy connections to the required address.
func proxyHandler() func(c echo.Context) error {
	const ProxyHandlerNetwork = "tcp"

	return func(c echo.Context) error {
		logger := log.WithFields(log.Fields{
			"remote":    c.Request().RemoteAddr,
			"namespace": c.Request().Header.Get("X-Namespace"),
			"path":      c.Request().Header.Get("X-Path"),
			"version":   AgentVersion,
		})

		errorResponse := func(err error, msg string, code int) error {
			logger.WithError(err).Debug(msg)

			return c.String(code, msg)
		}

		// NOTE: Gets the to address to connect to. This address can be just a port, :8080, or the host and port,
		// localhost:8080.
		addr := c.Param("addr")

		in, err := net.Dial(ProxyHandlerNetwork, addr)
		if err != nil {
			return errorResponse(err, "failed to connect to the server on device", http.StatusInternalServerError)
		}

		defer in.Close()

		// NOTE: Inform to the connection that the dial was successfully.
		if err := c.NoContent(http.StatusOK); err != nil {
			return errorResponse(err, "failed to send the ok status code back to server", http.StatusInternalServerError)
		}

		// NOTE: Hijacks the connection to control the data transferred to the client connected. This way, we don't
		// depend upon anything externally, only the data.
		out, _, err := c.Response().Hijack()
		if err != nil {
			return errorResponse(err, "failed to hijack connection", http.StatusInternalServerError)
		}

		defer out.Close() // nolint:errcheck

		wg := new(sync.WaitGroup)
		done := sync.OnceFunc(func() {
			defer in.Close()
			defer out.Close()

			logger.Trace("close called on in and out connections")
		})

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			io.Copy(in, out) //nolint:errcheck
		}()

		wg.Add(1)
		go func() {
			defer done()
			defer wg.Done()

			io.Copy(out, in) //nolint:errcheck
		}()

		logger.WithError(err).Trace("proxy handler waiting for data pipe")
		wg.Wait()

		logger.WithError(err).Trace("proxy handler done")

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

// Listen creates the SSH server and listening for connections.
func (a *Agent) Listen(ctx context.Context) error {
	a.mode.Serve(a)

	a.tunnel = tunnel.NewBuilder().
		WithConnHandler(connHandler(a.server)).
		WithCloseHandler(closeHandler(a, a.server)).
		WithProxyHandler(proxyHandler()).
		Build()

	go a.ping(ctx, AgentPingDefaultInterval) //nolint:errcheck

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			if a.isClosed() {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
				}).Info("Stopped listening for connections")

				cancel()

				return
			}

			namespace := a.authData.Namespace
			tenantName := a.authData.Name
			sshEndpoint := a.serverInfo.Endpoints.SSH

			sshid := strings.NewReplacer(
				"{namespace}", namespace,
				"{tenantName}", tenantName,
				"{sshEndpoint}", strings.Split(sshEndpoint, ":")[0],
			).Replace("{namespace}.{tenantName}@{sshEndpoint}")

			listener, err := a.cli.NewReverseListener(ctx, a.authData.Token, "/ssh/connection")
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

			{
				// NOTE: Tunnel'll only realize that it lost its connection to the ShellHub SSH when the next
				// "keep-alive" connection fails. As a result, it will take this interval to reconnect to its server.
				err := a.tunnel.Listen(listener)

				log.WithError(err).WithFields(log.Fields{
					"namespace":      namespace,
					"hostname":       tenantName,
					"server_address": a.config.ServerAddress,
					"ssh_server":     sshEndpoint,
					"sshid":          sshid,
				}).Info("Tunnel listener closed")

				listener.Close() // nolint:errcheck
			}

			a.listening <- false
		}
	}()

	<-ctx.Done()

	return a.Close()
}

// AgentPingDefaultInterval is the default time interval between ping on agent.
const AgentPingDefaultInterval = 10 * time.Minute

// ping sends an authorization request to the ShellHub server at each interval.
// A random value between 10 and [config.MaxRetryConnectionTimeout] seconds is added to the interval
// each time the ticker is executed.
//
// Ping only sends requests to the server if the agent is listening for connections. If the agent is not
// listening, the ping process will be stopped. When the interval is 0, the default value is 10 minutes.
func (a *Agent) ping(ctx context.Context, interval time.Duration) error {
	a.listening = make(chan bool)

	if interval == 0 {
		interval = AgentPingDefaultInterval
	}

	<-a.listening // NOTE: wait for the first connection to start to ping the server.
	ticker := time.NewTicker(interval)

	for {
		if a.isClosed() {
			return nil
		}

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
				}).Debug("Starting the ping interval to server")

				ticker.Reset(interval)
			} else {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"timestamp":      time.Now(),
				}).Debug("Stopped pinging server due listener status")

				ticker.Stop()
			}
		case <-ticker.C:
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

			randTimeout := time.Duration(rand.Intn(a.config.MaxRetryConnectionTimeout-10)+10) * time.Second //nolint:gosec
			ticker.Reset(interval + randTimeout)
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

// GetInfo gets information like the version and the enpoints for HTTP and SSH to ShellHub server.
func GetInfo(cfg *Config) (*models.Info, error) {
	cli, err := client.NewClient(cfg.ServerAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the HTTP client")
	}

	info, err := cli.GetInfo(AgentVersion)
	if err != nil {
		return nil, err
	}

	return info, nil
}
