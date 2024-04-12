package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type APIKeyStore interface {
	APIKeyCreate(ctx context.Context, APIKey *models.APIKey) error
	APIKeyList(ctx context.Context, tenantID string, paginator query.Paginator, sorter query.Sorter) ([]models.APIKey, int, error)
	APIKeyGetByUID(ctx context.Context, uid string) (*models.APIKey, error)
	APIKeyGetByName(ctx context.Context, tenandID string, name string) (*models.APIKey, error)
	APIKeyEdit(ctx context.Context, changes *requests.APIKeyChanges) error
	APIKeyDelete(ctx context.Context, id, tenantID string) error
}
