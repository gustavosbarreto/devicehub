package agent

import (
	"context"
	"os/exec"

	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/agent/server"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/classic"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/connector"
)

type Info struct {
	ID   string
	Name string
}

// Mode is the Agent execution mode.
//
// The Agent can be executed in two different modes: `Host` and `Connector`.
// The `Host` mode is the default one, where the agent will listen for incoming connections and use the host device as
// source of any information needed to start itself. When running in `Connector` mode, it uses the Docker engine as this
// source.
//
// Check [HostMode] and [ConnectorMode] for more information.
type Mode interface {
	// Serve prepares the Agent for listening, setting up the SSH server, its modes and values on Agent's.
	Serve(agent *Agent)
	// GetInfo gets information about Agent according to Agent's mode.
	//
	// When Agent is running on [HostMode], the info got is from the system where the Agent is running, but when running
	// in [ConnectorMode], the data is retrieved from Docker Engine.
	GetInfo() (*Info, error)
}

// ModeHost is the Agent execution mode for `Host`.
//
// The host mode is the default mode one, and turns the host machine into a ShellHub's Agent. The host is
// responsible for the SSH server, authentication and authorization, `/etc/passwd`, `/etc/shadow`, and etc.
type HostMode struct{}

var _ Mode = new(HostMode)

func (m *HostMode) Serve(agent *Agent) {
	agent.server = server.NewServer(
		agent.cli,
		agent.authData,
		agent.config.PrivateKey,
		agent.config.KeepAliveInterval,
		agent.config.SingleUserPassword,
		&classic.Mode{
			Authenticator: *classic.NewAuthenticator(agent.cli, agent.authData, agent.config.SingleUserPassword, &agent.authData.Name),
			Sessioner:     *classic.NewSessioner(&agent.authData.Name, make(map[string]*exec.Cmd)),
		},
	)

	agent.server.SetDeviceName(agent.authData.Name)
}

func (m *HostMode) GetInfo() (*Info, error) {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   osrelease.ID,
		Name: osrelease.Name,
	}, nil
}

// ModeConnector is the Agent execution mode for `Connector`.
//
// The `Connector` mode is used to turn a container inside a host into a single device ShellHub's Agent. The host is
// responsible for the SSH server, but the authentication and authorization is made by either the conainer
// internals, `passwd` or `shadow`, or by the ShellHub API.
type ConnectorMode struct {
	cli      *dockerclient.Client
	identity string
}

func NewConnectorMode(cli *dockerclient.Client, identity string) (Mode, error) {
	return &ConnectorMode{
		cli:      cli,
		identity: identity,
	}, nil
}

var _ Mode = new(ConnectorMode)

func (m *ConnectorMode) Serve(agent *Agent) {
	// NOTICE: When the agent is running in `Connector` mode, we need to identify the container ID to maintain the
	// communication between the server and the agent when the container name on the host changes.  This information is
	// saved inside the device's identity, avoiding significant changes in the current state of the agent.
	// TODO: Evaluate if we can use another field than "MAC" to store the container ID.
	agent.server = server.NewServer(
		agent.cli,
		agent.authData,
		agent.config.PrivateKey,
		agent.config.KeepAliveInterval,
		agent.config.SingleUserPassword,
		&connector.Mode{
			Authenticator: *connector.NewAuthenticator(agent.cli, m.cli, agent.authData, &agent.Identity.MAC),
			Sessioner:     *connector.NewSessioner(&agent.Identity.MAC, m.cli),
		},
	)

	agent.server.SetContainerID(agent.Identity.MAC)
	agent.server.SetDeviceName(agent.authData.Name)
}

func (m *ConnectorMode) GetInfo() (*Info, error) {
	info, err := m.cli.ContainerInspect(context.Background(), m.identity)
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   "docker",
		Name: info.Config.Image,
	}, nil
}
