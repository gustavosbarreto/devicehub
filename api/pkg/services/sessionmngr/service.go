package sessionmngr

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Service interface {
	CountSessions(ctx context.Context) (int64, error)
	ListSessions(ctx context.Context, perPage int, page int) ([]models.Session, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) CountSessions(ctx context.Context) (int64, error) {
	return s.store.CountSessions(ctx)
}

func (s *service) ListSessions(ctx context.Context, perPage int, page int) ([]models.Session, error) {
	return s.store.ListSessions(ctx, perPage, page)
}

func (s *service) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	return s.store.GetSession(ctx, uid)
}

func (s *service) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	return s.store.CreateSession(ctx, session)
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	return s.store.DeactivateSession(ctx, uid)
}

func (s *service) SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error {
	return s.store.SetSessionAuthenticated(ctx, uid, authenticated)
}
