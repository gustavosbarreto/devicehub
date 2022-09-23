package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/loglevel"

	"github.com/Masterminds/semver"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/agent/selfupdater"
	"github.com/shellhub-io/shellhub/agent/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option (e.g: `go build -ldflags "-X
// main.AgentVersion=1.2.3"`).
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

// ConfigOptions provides the configuration for the agent service. The values are load from
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

// NewAgentServer creates a new agent server instance.
func NewAgentServer() *Agent {
	opts := ConfigOptions{}

	// Process unprefixed env vars for backward compatibility
	envconfig.Process("", &opts) // nolint:errcheck

	if err := envconfig.Process("shellhub", &opts); err != nil {
		// show envconfig usage help users to run agent
		envconfig.Usage("shellhub", &opts) // nolint:errcheck
		log.Fatal(err)
	}

	// Set the log level accordingly to the configuration.
	level, err := log.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Error("Invalid log level has been provided.")
		os.Exit(1)
	}
	log.SetLevel(level)

	if os.Geteuid() == 0 && opts.SingleUserPassword != "" {
		log.Error("ShellHub agent cannot run as root when single-user mode is enabled.")
		log.Error("To disable single-user mode unset SHELLHUB_SINGLE_USER_PASSWORD env.")
		os.Exit(1)
	}

	if os.Geteuid() != 0 && opts.SingleUserPassword == "" {
		log.Error("When running as non-root user you need to set password for single-user mode by SHELLHUB_SINGLE_USER_PASSWORD environment variable.")
		log.Error("You can use openssl passwd utility to generate password hash. The following algorithms are supported: bsd1, apr1, sha256, sha512.")
		log.Error("Example: SHELLHUB_SINGLE_USER_PASSWORD=$(openssl passwd -6)")
		log.Error("See man openssl-passwd for more information.")
		os.Exit(1)
	}

	updater, err := selfupdater.NewUpdater(AgentVersion)
	if err != nil {
		log.Panic(err)
	}

	if err := updater.CompleteUpdate(); err != nil {
		log.Warning(err)
		os.Exit(0)
	}

	currentVersion := new(semver.Version)

	if AgentVersion != "latest" {
		currentVersion, err = updater.CurrentVersion()
		if err != nil {
			log.Panic(err)
		}
	}

	log.WithFields(log.Fields{
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
		log.Fatal(err)
	}

	if err := agent.initialize(); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to initialize agent")
	}

	serv := server.NewServer(agent.cli, agent.authData, opts.PrivateKey, opts.KeepAliveInterval, opts.SingleUserPassword)

	tun := tunnel.NewTunnel()
	tun.ConnHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		conn, ok := r.Context().Value("http-conn").(net.Conn)
		if !ok {
			log.WithFields(log.Fields{
				"version": AgentVersion,
			}).Warning("Type assertion failed")

			return
		}

		serv.Sessions[vars["id"]] = conn
		serv.HandleConn(conn)
	}
	tun.CloseHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serv.CloseSession(vars["id"])
	}

	serv.SetDeviceName(agent.authData.Name)

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

			log.WithFields(log.Fields{
				"namespace":      namespace,
				"hostname":       tenantName,
				"server_address": opts.ServerAddress,
				"ssh_server":     sshEndpoint,
				"sshid":          sshid,
			}).Info("Server connection established")

			if err := tun.Listen(listener); err != nil {
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
					log.Error(err)

					goto sleep
				}

				if nextVersion.GreaterThan(currentVersion) {
					if err := updater.ApplyUpdate(nextVersion); err != nil {
						log.Error(err)
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
		sessions := make([]string, 0, len(serv.Sessions))
		for key := range serv.Sessions {
			sessions = append(sessions, key)
		}

		agent.sessions = sessions

		if err := agent.authorize(); err != nil {
			serv.SetDeviceName(agent.authData.Name)
		}
	}

	return agent
}

func main() {
	// Default command.
	rootCmd := &cobra.Command{ // nolint: exhaustruct
		Use: "agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			NewAgentServer()
		},
	}

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "info",
		Short: "Show information about the agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			if err := NewAgentServer().probeServerInfo(); err != nil {
				log.Fatal(err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "sftp",
		Short: "Starts the SFTP server",
		Long: `Starts the SFTP server. This command is used internally by the agent and should not be used directly.
It is initialized by the agent when a new SFTP session is created.`,
		Run: func(cmd *cobra.Command, args []string) {
			NewSFTPServer()
		},
	})

	rootCmd.Execute() // nolint: errcheck
}
