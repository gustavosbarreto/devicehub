package connector

import (
	"context"
)

// ConnectorVersion stores the version of the ShellHub Instane that is running the connector.
// It is used in the ShellHub Agents initialized by the connector when a container is started.
var ConnectorVersion string

// Container is a struct that represents a container that will be managed by the connector.
type Container struct {
	// ID is the container ID.
	ID string
	// Name is the container name.
	Name string
	// ServerAddress is the ShellHub address of the server that the agent will connect to.
	ServerAddress string
	// Tenant is the tenant ID of the namespace that the agent belongs to.
	Tenant string
	// PrivateKey is the private key of the device. Specify the path to store the container private key. If not
	// provided, the agent will generate a new one. This is required.
	PrivateKey string
	// Cancel is a function that is used to stop the goroutine that is running the agent for this container.
	Cancel context.CancelFunc
}

// Connector is an interface that defines the methods that a connector must implement.
type Connector interface {
	// List lists all containers running on the host.
	List(ctx context.Context) ([]Container, error)
	// Start starts the agent for the container with the given ID.
	Start(ctx context.Context, id string, name string)
	// Stop stops the agent for the container with the given ID.
	Stop(ctx context.Context, id string)
	// Listen listens for events and starts or stops the agent for the container that was created or removed.
	Listen(ctx context.Context) error
}
