package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/Masterminds/semver"
	"github.com/shellhub-io/shellhub/pkg/agent"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/selfupdater"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/connector"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
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

func main() {
	// Default command.
	rootCmd := &cobra.Command{ // nolint: exhaustruct
		Use: "agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			cfg, fields, err := agent.LoadConfigFromEnv()
			if err != nil {
				log.WithError(err).WithFields(fields).Fatal("Failed to load de configuration from the environmental variables")
			}

			if os.Geteuid() == 0 && cfg.SingleUserPassword != "" {
				log.Error("ShellHub agent cannot run as root when single-user mode is enabled.")
				log.Error("To disable single-user mode unset SHELLHUB_SINGLE_USER_PASSWORD env.")
				os.Exit(1)
			}

			if os.Geteuid() != 0 && cfg.SingleUserPassword == "" {
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

			mode := func() string {
				if cfg.SingleUserPassword != "" {
					return "single-user"
				}

				return "multi-user"
			}()

			log.WithFields(log.Fields{
				"version": AgentVersion,
				"mode":    mode,
			}).Info("Starting ShellHub")

			ag, err := agent.NewAgentWithConfig(cfg, new(agent.HostMode))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to create agent")
			}

			if err := ag.Initialize(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to initialize agent")
			}

			ctx := cmd.Context()

			log.WithFields(log.Fields{
				"version":            AgentVersion,
				"mode":               mode,
				"tenant_id":          cfg.TenantID,
				"server_address":     cfg.ServerAddress,
				"preferred_hostname": cfg.PreferredHostname,
			}).Info("Listening for SSH connections")

			// Disable check update in development mode
			if AgentVersion != "latest" {
				go func() {
					for {
						nextVersion, err := ag.CheckUpdate()
						if err != nil {
							log.WithError(err).WithFields(log.Fields{
								"version":            AgentVersion,
								"mode":               mode,
								"tenant_id":          cfg.TenantID,
								"server_address":     cfg.ServerAddress,
								"preferred_hostname": cfg.PreferredHostname,
							}).Error("Failed to check update")

							goto sleep
						}

						if nextVersion.GreaterThan(currentVersion) {
							if err := updater.ApplyUpdate(nextVersion); err != nil {
								log.WithError(err).WithFields(log.Fields{
									"version":            AgentVersion,
									"mode":               mode,
									"tenant_id":          cfg.TenantID,
									"server_address":     cfg.ServerAddress,
									"preferred_hostname": cfg.PreferredHostname,
								}).Error("Failed to apply update")
							}

							log.WithFields(log.Fields{
								"version":            currentVersion,
								"next_version":       nextVersion.String(),
								"mode":               mode,
								"tenant_id":          cfg.TenantID,
								"server_address":     cfg.ServerAddress,
								"preferred_hostname": cfg.PreferredHostname,
							}).Info("Update successfully applied")
						}

					sleep:
						log.WithFields(log.Fields{
							"version":            AgentVersion,
							"mode":               mode,
							"tenant_id":          cfg.TenantID,
							"server_address":     cfg.ServerAddress,
							"preferred_hostname": cfg.PreferredHostname,
						}).Info("Sleeping for 24 hours")

						time.Sleep(time.Hour * 24)
					}
				}()
			}

			go func() {
				if err := ag.ListenSSH(ctx); err != nil {
					log.WithError(err).WithFields(log.Fields{
						"version":            AgentVersion,
						"mode":               mode,
						"tenant_id":          cfg.TenantID,
						"server_address":     cfg.ServerAddress,
						"preferred_hostname": cfg.PreferredHostname,
					}).Fatal("Failed to listen for SSH connections")
				}
			}()

			go func() {
				if !cfg.VPN {
					log.Info("VPN is disable")

					return
				}

				log.Debug("VPN enabled")

				for {
					log.Info("VPN connected")

					if err := ag.ConnectVPN(ctx); err != nil {
						log.WithError(err).Error("Connect to VPN lost. Retrying in 10 seconds.")
					}

					time.Sleep(10 * time.Second)
				}
			}()

			<-ctx.Done()

			log.WithFields(log.Fields{
				"version":            AgentVersion,
				"mode":               mode,
				"tenant_id":          cfg.TenantID,
				"server_address":     cfg.ServerAddress,
				"preferred_hostname": cfg.PreferredHostname,
			}).Info("Agent Stopped")
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "connector",
		Short: "Starts the ShellHub Agent in Connector mode",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, fields, err := connector.LoadConfigFromEnv()
			if err != nil {
				log.WithError(err).
					WithFields(fields).
					Fatal("Failed to load de configuration from the environmental variables")
			}

			logger := log.WithFields(
				log.Fields{
					"address":      cfg.ServerAddress,
					"tenant_id":    cfg.TenantID,
					"private_keys": cfg.PrivateKeys,
					"version":      AgentVersion,
				},
			)

			cfg.PrivateKeys = path.Dir(cfg.PrivateKeys)

			logger.Info("Starting ShellHub Agent Connector")

			connector.ConnectorVersion = AgentVersion
			connector, err := connector.NewDockerConnector(cfg.ServerAddress, cfg.TenantID, cfg.PrivateKeys)
			if err != nil {
				logger.Fatal("Failed to create ShellHub Agent Connector")
			}

			if err := connector.Listen(cmd.Context()); err != nil {
				logger.Fatal("Failed to listen for connections")
			}

			logger.Info("ShellHub Agent Connector stopped")
		},
	})

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "info",
		Short: "Show information about the agent",
		Run: func(cmd *cobra.Command, args []string) {
			loglevel.SetLogLevel()

			cfg, err := envs.ParseWithPrefix[agent.Config]("SHELLHUB_")
			if err != nil {
				log.Fatal(err)
			}

			info, err := agent.GetInfo(cfg)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to get agent information")
			}

			log.WithFields(log.Fields{
				"version": info.Version,
				"api":     info.Endpoints.API,
				"ssh":     info.Endpoints.SSH,
			}).Info("ShellHub agent information")

			data, err := json.Marshal(info)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"version":       AgentVersion,
					"configuration": cfg,
				}).Fatal("Failed to marshal agent information")
			}

			// NOTICE: this output was made to enable the agent's user to check and parse the agent's information with
			// a know format without having to parse the log output.
			// TODO: Should it have line break or not?
			cmd.Println(string(data))
		},
	})

	rootCmd.AddCommand(&cobra.Command{ // nolint: exhaustruct
		Use:   "sftp",
		Short: "Starts the SFTP server",
		Long: `Starts the SFTP server. This command is used internally by the agent and should not be used directly.
It is initialized by the agent when a new SFTP session is created.`,
		Run: func(cmd *cobra.Command, args []string) {
			ssh.NewSFTPServer(command.SFTPServerMode(args[0]))
		},
	})

	rootCmd.Version = AgentVersion

	rootCmd.SetVersionTemplate(fmt.Sprintf("{{ .Name }} version: {{ .Version }}\ngo: %s\n",
		runtime.Version(),
	))

	agent.AgentVersion = AgentVersion
	agent.AgentPlatform = AgentPlatform

	rootCmd.Execute() // nolint: errcheck
}
