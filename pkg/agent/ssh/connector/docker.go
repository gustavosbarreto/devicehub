package connector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/agent"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/validator"
	log "github.com/sirupsen/logrus"
)

var _ Connector = new(DockerConnector)

// DockerConnector is a struct that represents a connector that uses Docker as the container runtime.
type DockerConnector struct {
	mu sync.Mutex
	// server is the ShellHub address of the server that the agent will connect to.
	server string
	// tenant is the tenant ID of the namespace that the agent belongs to.
	tenant string
	// cli is the Docker client.
	cli *dockerclient.Client
	// privateKeys is the path to the directory that contains the private keys for the containers.
	privateKeys string
	// cancels is a map that contains the cancel functions for each container.
	// This is used to stop the agent for a container, marking as done its context and closing the agent.
	cancels map[string]context.CancelFunc
}

// Config provides the configuration for the agent connector service.
type Config struct {
	// Set the ShellHub server address the agent will use to connect.
	// This is required.
	ServerAddress string `env:"SERVER_ADDRESS,required"`

	// Specify the path to store the devices/containers private keys.
	// If not provided, the agent will generate a new one.
	// This is required.
	PrivateKeys string `env:"PRIVATE_KEYS,required"`

	// Sets the account tenant id used during communication to associate the
	// devices to a specific tenant.
	// This is required.
	TenantID string `env:"TENANT_ID,required"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval int `env:"KEEPALIVE_INTERVAL,overwrite,default=30"`
}

func LoadConfigFromEnv() (*Config, map[string]interface{}, error) {
	cfg, err := envs.ParseWithPrefix[Config]("SHELLHUB_")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: test the envinromental variables validation on integration tests.
	if ok, fields, err := validator.New().StructWithFields(cfg); err != nil || !ok {
		log.WithFields(fields).Error("failed to validate the configuration loaded from envs")

		return nil, fields, err
	}

	return cfg, nil, nil
}

func NewDockerConnectorWithClient(cli *dockerclient.Client, server string, tenant string, privateKey string) (Connector, error) {
	return &DockerConnector{
		server:      server,
		tenant:      tenant,
		cli:         cli,
		privateKeys: privateKey,
		cancels:     make(map[string]context.CancelFunc),
	}, nil
}

// NewDockerConnector creates a new [Connector] that uses Docker as the container runtime.
func NewDockerConnector(server string, tenant string, privateKey string) (Connector, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerConnector{
		server:      server,
		tenant:      tenant,
		cli:         cli,
		privateKeys: privateKey,
		cancels:     make(map[string]context.CancelFunc),
	}, nil
}

// events returns the docker events.
func (d *DockerConnector) events(ctx context.Context) (<-chan events.Message, <-chan error) {
	return d.cli.Events(ctx, events.ListOptions{})
}

func (d *DockerConnector) List(ctx context.Context) ([]Container, error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	list := make([]Container, len(containers))
	for i, container := range containers {
		list[i].ID = container.ID

		name, err := d.getContainerNameFromID(ctx, container.ID)
		if err != nil {
			return nil, err
		}

		list[i].Name = name
	}

	return list, nil
}

// Start starts the agent for the container with the given ID.
func (d *DockerConnector) Start(ctx context.Context, id string, name string) {
	id = id[:12]

	d.mu.Lock()
	ctx, d.cancels[id] = context.WithCancel(ctx)
	d.mu.Unlock()

	privateKey := fmt.Sprintf("%s/%s.key", d.privateKeys, id)
	go initContainerAgent(ctx, d.cli, Container{
		ID:            id,
		Name:          name,
		ServerAddress: d.server,
		Tenant:        d.tenant,
		PrivateKey:    privateKey,
		Cancel:        d.cancels[id],
	})
}

// Stop stops the agent for the container with the given ID.
func (d *DockerConnector) Stop(_ context.Context, id string) {
	id = id[:12]

	d.mu.Lock()
	defer d.mu.Unlock()

	cancel, ok := d.cancels[id]
	if ok {
		cancel()
		delete(d.cancels, id)
	}
}

func (d *DockerConnector) getContainerNameFromID(ctx context.Context, id string) (string, error) {
	container, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", err
	}

	// NOTICE: It removes the first character on container's name that is a `/`.
	return container.Name[1:], nil
}

// Listen listens for events and starts or stops the agent for the containers.
func (d *DockerConnector) Listen(ctx context.Context) error {
	containers, err := d.List(ctx)
	if err != nil {
		return err
	}

	for _, container := range containers {
		d.Start(ctx, container.ID, container.Name)
	}

	events, errs := d.events(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			return err
		case container := <-events:
			// NOTICE: "start" and "die" Docker's events are call every time a new container start or stop,
			// independently how the command was run. For example, if a container was started with `docker run -d`, the
			// "start" event will be called, but if the same container was started with `docker start <container-id>`,
			// the "start" event will be called too. The same happens with the "die" event.
			switch container.Action {
			case "start":
				name, err := d.getContainerNameFromID(ctx, container.ID)
				if err != nil {
					return err
				}

				d.Start(ctx, container.ID, name)
			case "die":
				d.Stop(ctx, container.ID)
			}
		}
	}
}

// initContainerAgent initializes the agent for a container.
func initContainerAgent(ctx context.Context, cli *dockerclient.Client, container Container) {
	agent.AgentPlatform = "connector"
	agent.AgentVersion = ConnectorVersion

	// TODO: Let this configuration build next to the Agent [agent.LoadConfigFromEnv] function.
	cfg := &agent.Config{
		ServerAddress:             container.ServerAddress,
		TenantID:                  container.Tenant,
		PrivateKey:                container.PrivateKey,
		PreferredIdentity:         container.ID,
		PreferredHostname:         container.Name,
		KeepAliveInterval:         30,
		MaxRetryConnectionTimeout: 60,
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"timestamp":      time.Now(),
		"version":        agent.AgentVersion,
	}).Info("Connector container started")

	mode, err := agent.NewConnectorMode(cli, container.ID)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"timestamp":      time.Now(),
			"version":        agent.AgentVersion,
		}).Fatal("Failed to create connector mode")
	}

	ag, err := agent.NewAgentWithConfig(cfg, mode)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":            container.ID,
			"configuration": cfg,
			"version":       agent.AgentVersion,
		}).Fatal("Failed to create agent")
	}

	if err := ag.Initialize(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":            container.ID,
			"configuration": cfg,
			"version":       agent.AgentVersion,
		}).Fatal("Failed to initialize agent")
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"timestamp":      time.Now(),
		"version":        agent.AgentVersion,
	}).Info("Listening for connections")

	// NOTICE(r): listing for connection and wait for a channel message to close the agent. It will receives
	// this mensagem when something out of this goroutine send a `done`, what will cause the agent closes
	// and no more connection to be allowed until it be started again.
	if err := ag.ListenSSH(ctx); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"timestamp":      time.Now(),
			"version":        agent.AgentVersion,
		}).Fatal("Failed to listen for connections")
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"version":        agent.AgentVersion,
	}).Info("Connector container done")
}
