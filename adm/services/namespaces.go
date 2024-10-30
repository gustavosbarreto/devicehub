package services

import (
	"context"

	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// NamespaceCreate initializes a new namespace, making the specified user its owner.
// The tenant defaults to a UUID if not provided.
// Max device limit is based on the envs.IsCloud() setting.
func (s *service) NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error) {
	// tenant is optional.
	if input.TenantID == "" {
		input.TenantID = uuid.Generate()
	}

	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrNamespaceInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Owner)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns := &models.Namespace{
		Name:       input.Namespace,
		Owner:      user.ID,
		TenantID:   input.TenantID,
		MaxDevices: getMaxDevices(),
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				AddedAt: clock.Now(),
				Status:  models.MemberStatusAccepted,
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: models.DefaultAnnouncementMessage,
		},
		CreatedAt: clock.Now(),
	}

	ns, err = s.store.NamespaceCreate(ctx, ns)
	if err != nil {
		return nil, ErrDuplicateNamespace
	}

	return ns, nil
}

// NamespaceAddMember adds a new member with a specified role to a namespace.
func (s *service) NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error) {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrInvalidFormat
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(ctx, input.Namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	if err = s.store.NamespaceAddMember(ctx, ns.TenantID, &models.Member{ID: user.ID, Role: input.Role}); err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return ns, nil
}

// NamespaceRemoveMember removes a member from a namespace.
func (s *service) NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error) {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrInvalidFormat
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(ctx, input.Namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	if err = s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID); err != nil {
		return nil, ErrFailedNamespaceRemoveMember
	}

	return ns, nil
}

// NamespaceDelete deletes a namespace based on the provided namespace name.
func (s *service) NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrNamespaceInvalid
	}

	ns, err := s.store.NamespaceGetByName(ctx, input.Namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(ctx, ns.TenantID); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}
