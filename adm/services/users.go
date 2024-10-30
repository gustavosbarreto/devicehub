package services

import (
	"context"
	"slices"

	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// UserCreate adds a new user based on the provided user's data. This method validates data and
// checks for conflicts.
func (s *service) UserCreate(ctx context.Context, input *inputs.UserCreate) (*models.User, error) {
	// TODO: convert username and email to lower case.
	userData := models.UserData{
		Name:     input.Username,
		Email:    input.Email,
		Username: input.Username,
	}

	// TODO: validate this at cmd layer
	if ok, err := s.validator.Struct(userData); !ok || err != nil {
		return nil, ErrUserDataInvalid
	}

	if conflicts, has, _ := s.store.UserConflicts(ctx, &models.UserConflicts{Email: userData.Email, Username: userData.Username}); has {
		containsEmail := slices.Contains(conflicts, "email")
		containsUsername := slices.Contains(conflicts, "username")

		switch {
		case containsUsername && containsEmail:
			return nil, ErrUserNameAndEmailExists
		case containsUsername:
			return nil, ErrUserNameExists
		case containsEmail:
			return nil, ErrUserEmailExists
		default:
			return nil, ErrUserUnhandledDuplicate
		}
	}

	password, err := models.HashUserPassword(input.Password)
	if err != nil {
		return nil, ErrUserPasswordInvalid
	}

	// TODO: validate this at cmd layer
	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return nil, ErrUserPasswordInvalid
	}

	user := &models.User{
		UserData:      userData,
		Password:      password,
		Status:        models.UserStatusConfirmed,
		CreatedAt:     clock.Now(),
		MaxNamespaces: MaxNumberNamespacesCommunity,
	}

	if _, err := s.store.UserCreate(ctx, user); err != nil {
		return nil, ErrCreateNewUser
	}

	return user, nil
}

// UserDelete removes a user and cleans up related data based on the provided username.
func (s *service) UserDelete(ctx context.Context, input *inputs.UserDelete) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	userInfo, err := s.store.UserGetInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}

	for _, ns := range userInfo.OwnedNamespaces {
		if err := s.store.NamespaceDelete(ctx, ns.TenantID); err != nil {
			return err
		}
	}

	for _, ns := range userInfo.AssociatedNamespaces {
		if err := s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID); err != nil {
			return err
		}
	}

	if err := s.store.UserDelete(ctx, user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

// UserUpdate updates a user's data based on the provided username.
func (s *service) UserUpdate(ctx context.Context, input *inputs.UserUpdate) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrUserDataInvalid
	}

	user, err := s.store.UserGetByUsername(ctx, input.Username)
	if err != nil {
		return ErrUserNotFound
	}

	password, err := models.HashUserPassword(input.Password)
	if err != nil {
		return ErrUserPasswordInvalid
	}

	// TODO: validate this at cmd layer
	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return ErrUserPasswordInvalid
	}

	if err := s.store.UserUpdate(ctx, user.ID, &models.UserChanges{Password: password.Hash}); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
