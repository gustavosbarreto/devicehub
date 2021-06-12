package nsadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserNotFound      = errors.New("user not found")
	ErrNamespaceNotFound = errors.New("namespace not found")
	ErrDuplicateID       = errors.New("user already member of this namespace")
	ErrConflictName      = errors.New("this name already exists")
	ErrInvalidFormat     = errors.New("invalid name format")
)

type Service interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerUsername string) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, tenantID, ownerUsername string) error
	EditNamespace(ctx context.Context, tenantID, name, ownerUsername string) (*models.Namespace, error)
	AddNamespaceUser(ctx context.Context, tenantID, username, ownerUsername string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, tenantID, username, ownerUsername string) (*models.Namespace, error)
	ListMembers(ctx context.Context, tenantID string) ([]models.Member, error)
	UpdateDataUserSecurity(ctx context.Context, status bool, tenant string) error
	GetDataUserSecurity(ctx context.Context, tenant string) (bool, error)
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter

	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.NamespaceList(ctx, pagination, filter, export)
}

func (s *service) CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerID string) (*models.Namespace, error) {
	user, _, err := s.store.UserGetByID(ctx, ownerID, false)
	if user == nil {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}

	ns := &models.Namespace{
		Name:     strings.ToLower(namespace.Name),
		Owner:    user.ID,
		Members:  []interface{}{user.ID},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: namespace.TenantID,
	}

	if namespace.TenantID == "" {
		ns.TenantID = uuid.Generate()
	}

	if envs.IsEnterprise() {
		ns.MaxDevices = 3
	} else {
		ns.MaxDevices = -1
	}

	otherNamespace, err := s.store.NamespaceGetByName(ctx, ns.Name)
	if err != nil && err != store.ErrNoDocuments {
		return nil, err
	}

	if otherNamespace != nil {
		return nil, ErrConflictName
	}

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

func (s *service) GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error) {
	return s.store.NamespaceGet(ctx, tenantID)
}

func (s *service) DeleteNamespace(ctx context.Context, tenantID, ownerID string) error {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return ErrNamespaceNotFound
	}

	if err != nil {
		return err
	}

	user, _, err := s.store.UserGetByID(ctx, ownerID, false)
	if err == store.ErrNoDocuments {
		return ErrUnauthorized
	}

	if err != nil {
		return err
	}

	if ns.Owner != user.ID {
		return ErrUnauthorized
	}

	return s.store.NamespaceDelete(ctx, tenantID)
}

func (s *service) ListMembers(ctx context.Context, tenantID string) ([]models.Member, error) {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return nil, ErrNamespaceNotFound
	}

	if err != nil {
		return nil, err
	}

	members := []models.Member{}
	for _, memberID := range ns.Members {
		user, _, err := s.store.UserGetByID(ctx, memberID.(string), false)
		if err == store.ErrNoDocuments {
			return nil, ErrUserNotFound
		}

		if err != nil {
			return nil, err
		}

		member := models.Member{ID: memberID.(string), Name: user.Username}
		members = append(members, member)
	}

	return members, nil
}

func (s *service) EditNamespace(ctx context.Context, tenantID, name, owner string) (*models.Namespace, error) {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return nil, ErrNamespaceNotFound
	}

	if err != nil {
		return nil, err
	}

	user, _, err := s.store.UserGetByID(ctx, owner, false)
	if err == store.ErrNoDocuments {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}

	validate := validator.New()
	lowerName := strings.ToLower(name)
	if err := validate.Struct(ns); err != nil {
		return nil, ErrInvalidFormat
	}

	if ns.Name == lowerName || ns.Owner != user.ID {
		return nil, ErrUnauthorized
	}

	return s.store.NamespaceRename(ctx, ns.TenantID, lowerName)
}

func (s *service) AddNamespaceUser(ctx context.Context, tenantID, username, ownerID string) (*models.Namespace, error) {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return nil, ErrNamespaceNotFound
	}

	if err != nil {
		return nil, err
	}

	ownerUser, _, err := s.store.UserGetByID(ctx, ownerID, false)
	if err == store.ErrNoDocuments {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}

	if ns.Owner != ownerUser.ID {
		return nil, ErrUserNotFound
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err == store.ErrNoDocuments {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return s.store.NamespaceAddMember(ctx, tenantID, user.ID)
}

func (s *service) RemoveNamespaceUser(ctx context.Context, tenantID, username, ownerID string) (*models.Namespace, error) {
	if _, err := s.store.NamespaceGet(ctx, tenantID); err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrNamespaceNotFound
		}

		return nil, err
	}

	if _, _, err := s.store.UserGetByID(ctx, ownerID, false); err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUnauthorized
		}

		return nil, err
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err == store.ErrNoDocuments {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return s.store.NamespaceRemoveMember(ctx, tenantID, user.ID)
}

func (s *service) UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error {
	if _, err := s.GetNamespace(ctx, tenant); err != nil {
		if err == store.ErrNoDocuments {
			return ErrUnauthorized
		}

		return err
	}

	return s.store.NamespaceSetSessionRecord(ctx, sessionRecord, tenant)
}

func (s *service) GetDataUserSecurity(ctx context.Context, tenant string) (bool, error) {
	if _, err := s.store.NamespaceGet(ctx, tenant); err != nil {
		if err == store.ErrNoDocuments {
			return false, ErrUnauthorized
		}

		return false, err
	}

	return s.store.NamespaceGetSessionRecord(ctx, tenant)
}
