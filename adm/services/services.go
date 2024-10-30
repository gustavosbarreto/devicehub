package services

import (
	"context"

	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

const (
	MaxNumberNamespacesCommunity = -1 // MaxNumberNamespacesCommunity defines a unlimited number of namespaces for communitty environment.
	MaxNumberDevicesLimited      = 3  // MaxNumberDevicesLimited defines the limit of devices for cloud environment.
	MaxNumberDevicesUnlimited    = -1 // MaxNumberDevicesUnlimited defines a unlimited number of devices for enterprise and community environment.
)

type Services interface {
	// UserCreate adds a new user based on the provided user's data. This method validates data and
	// checks for conflicts.
	UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error)
	// UserDelete removes a user and cleans up related data based on the provided username.
	UserDelete(ctx context.Context, input *inputs.UserDelete) error
	// UserUpdate updates a user's data based on the provided username.
	UserUpdate(ctx context.Context, input *inputs.UserUpdate) error
	// NamespaceCreate initializes a new namespace, making the specified user its owner.
	// The tenant defaults to a UUID if not provided.
	// Max device limit is based on the envs.IsCloud() setting.
	NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error)
	// NamespaceDelete deletes a namespace based on the provided namespace name.
	NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error
	// NamespaceAddMember adds a new member with a specified role to a namespace.
	NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error)
	// NamespaceRemoveMember removes a member from a namespace.
	NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error)
}

// service is an internal struct that implements the Services interface.
// It contains a store, which provides a mechanism to interact with the data store.
type service struct {
	store     store.Store
	validator *validator.Validator
}

// NewService creates and returns a new instance of the service with the provided store.
func NewService(store store.Store) Services {
	return &service{store, validator.New()}
}
