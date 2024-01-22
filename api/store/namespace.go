package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceStore interface {
	NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, export bool) ([]models.Namespace, int, error)
	NamespaceGet(ctx context.Context, tenantID string) (*models.Namespace, error)
	NamespaceGetByName(ctx context.Context, name string) (*models.Namespace, error)
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	NamespaceRename(ctx context.Context, tenantID string, name string) (*models.Namespace, error)
	NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error
	NamespaceDelete(ctx context.Context, tenantID string) error
	NamespaceAddMember(ctx context.Context, tenantID string, memberID string, memberRole string) (*models.Namespace, error)
	NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) (*models.Namespace, error)
	NamespaceEditMember(ctx context.Context, tenantID string, memberID string, memberNewRole string) error
	NamespaceGetFirst(ctx context.Context, id string) (*models.Namespace, error)
	NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error
	NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error)
}
