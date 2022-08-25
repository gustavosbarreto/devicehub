package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/agent/selfupdater"
	"github.com/shellhub-io/shellhub/agent/sshd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option (e.g: `go build -ldflags "-X
// main.AgentVersion=1.2.3"`).
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

// Provides the configuration for the agent service. The values are load from
// the system environment and control multiple aspects of the service.
type ConfigOptions struct {
	// Set the ShellHub Cloud server address the agent will use to connect.
	ServerAddress string `envconfig:"server_address" required:"true"`

	// Specify the path to the device private key.
	PrivateKey string `envconfig:"private_key" required:"true"`

	// Sets the account tenant id used during communication to associate the
	// device to a specific tenant.
	TenantID string `envconfig:"tenant_id" required:"true"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval int `envconfig:"keepalive_interval" default:"30"`

	// Set the device preferred hostname. This provides a hint to the server to
	// use this as hostname if it is available.
	PreferredHostname string `envconfig:"preferred_hostname"`

	// Set the device preferred identity. This provides a hint to the server to
	// use this identity if it is available.
	PreferredIdentity string `envconfig:"preferred_identity" default:""`

	// Set password for single-user mode (without root privileges). If not provided,
	// multi-user mode (with root privileges) is enabled by default.
	// NOTE: The password hash could be generated by ```openssl passwd```.
	SingleUserPassword string `envconfig:"simple_user_password"`

	// Log level to use. Valid values are 'info', 'warning', 'error', 'debug', and 'trace'.
	LogLevel string `envconfig:"log_level" default:"info"`
}

func NewAgentServer() {
	opts := ConfigOptions{}

	// Process unprefixed env vars for backward compatibility
	envconfig.Process("", &opts) // nolint:errcheck

	if err := envconfig.Process("shellhub", &opts); err != nil {
		// show envconfig usage help users to run agent
		envconfig.Usage("shellhub", &opts) // nolint:errcheck
		logrus.Fatal(err)
	}

	// Set the log level accordingly to the configuration.
	level, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logrus.Error("Invalid log level has been provided.")
		os.Exit(1)
	}
	logrus.SetLevel(level)

	if os.Geteuid() == 0 && opts.SingleUserPassword != "" {
		logrus.Error("ShellHub agent cannot run as root when single-user mode is enabled.")
		logrus.Error("To disable single-user mode unset SHELLHUB_SINGLE_USER_PASSWORD env.")
		os.Exit(1)
	}

	if os.Geteuid() != 0 && opts.SingleUserPassword == "" {
		logrus.Error("When running as non-root user you need to set password for single-user mode by SHELLHUB_SINGLE_USER_PASSWORD environment variable.")
		logrus.Error("You can use openssl passwd utility to generate password hash. The following algorithms are supported: bsd1, apr1, sha256, sha512.")
		logrus.Error("Example: SHELLHUB_SINGLE_USER_PASSWORD=$(openssl passwd -6)")
		logrus.Error("See man openssl-passwd for more information.")
		os.Exit(1)
	}

	updater, err := selfupdater.NewUpdater(AgentVersion)
	if err != nil {
		logrus.Panic(err)
	}

	if err := updater.CompleteUpdate(); err != nil {
		logrus.Warning(err)
		os.Exit(0)
	}

	currentVersion := new(semver.Version)

	if AgentVersion != "latest" {
		currentVersion, err = updater.CurrentVersion()
		if err != nil {
			logrus.Panic(err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"version": AgentVersion,
		"mode": func() string {
			if opts.SingleUserPassword != "" {
				return "single-user"
			}

			return "multi-user"
		}(),
	}).Info("Starting ShellHub")

	agent, err := NewAgent(&opts)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := agent.initialize(); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to initialize agent")
	}

	sshserver := sshd.NewServer(agent.cli, agent.authData, opts.PrivateKey, opts.KeepAliveInterval, opts.SingleUserPassword)

	tunnel := NewTunnel()
	tunnel.connHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		conn, ok := r.Context().Value("http-conn").(net.Conn)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"version": AgentVersion,
			}).Warning("Type assertion failed")

			return
		}

		sshserver.Sessions[vars["id"]] = conn
		sshserver.HandleConn(conn)
	}
	tunnel.closeHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sshserver.CloseSession(vars["id"])
	}

	sshserver.SetDeviceName(agent.authData.Name)

	go func() {
		for {
			listener, err := agent.newReverseListener()
			if err != nil {
				time.Sleep(time.Second * 10)

				continue
			}

			namespace := agent.authData.Namespace
			tenantName := agent.authData.Name
			sshEndpoint := agent.serverInfo.Endpoints.SSH

			sshid := strings.NewReplacer(
				"{namespace}", namespace,
				"{tenantName}", tenantName,
				"{sshEndpoint}", strings.Split(sshEndpoint, ":")[0],
			).Replace("{namespace}.{tenantName}@{sshEndpoint}")

			logrus.WithFields(logrus.Fields{
				"namespace":      namespace,
				"hostname":       tenantName,
				"server_address": opts.ServerAddress,
				"ssh_server":     sshEndpoint,
				"sshid":          sshid,
			}).Info("Server connection established")

			if err := tunnel.Listen(listener); err != nil {
				continue
			}
		}
	}()

	// Disable check update in development mode
	if AgentVersion != "latest" {
		go func() {
			for {
				nextVersion, err := agent.checkUpdate()
				if err != nil {
					logrus.Error(err)

					goto sleep
				}

				if nextVersion.GreaterThan(currentVersion) {
					if err := updater.ApplyUpdate(nextVersion); err != nil {
						logrus.Error(err)
					}
				}

			sleep:
				time.Sleep(time.Hour * 24)
			}
		}()
	}

	// This hard coded interval will be removed in a follow up change to make use of JWT token expire time.
	ticker := time.NewTicker(10 * time.Minute)

	for range ticker.C {
		sessions := make([]string, 0, len(sshserver.Sessions))
		for key := range sshserver.Sessions {
			sessions = append(sessions, key)
		}

		agent.sessions = sessions

		if err := agent.authorize(); err != nil {
			sshserver.SetDeviceName(agent.authData.Name)
		}
	}

	rootCmd := &cobra.Command{Use: "agent"}

	rootCmd.AddCommand(&cobra.Command{
		Use: "info",
		Run: func(cmd *cobra.Command, args []string) {
			if err := agent.probeServerInfo(); err != nil {
				logrus.Fatal(err)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "sftp" {
		NewSFTPServer()
	} else {
		NewAgentServer()
	}
}
