package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type TagsStore interface {
	// TagsGet retrieves all tags associated with the specified tenant. It functions by invoking "[document]GetTags"
	// for each document that implements tags.
	// Returns the tags, the count of unique tags, and an error if any issues arise.
	// It also filters the returned tags, removing any duplicates.
	TagsGet(ctx context.Context, tenant string) (tags []models.Tags, n int64, err error)

	// TagsRename replaces all occurrences of the old tag with the new tag for all documents associated with the specified tenant.
	// It operates by invoking "[document]BulkRenameTag" for each document that implements tags.
	// Returns the count of documents updated and an error if any issues arise during the tag renaming.
	TagsRename(ctx context.Context, tenant string, oldTag string, newTag string) (updatedCount int64, err error)

	// TagsDelete removes a tag from all documents associated with the specified tenant. It operates by
	// invoking "[document]BulkDeleteTag" for each document that implements tags.
	// Returns the count of documents updated and an error if any issues arise during the tag deletion.
	TagsDelete(ctx context.Context, tenant string, tag string) (updatedCount int64, err error)

	TagGet(ctx context.Context, tagName, tenant string) (*models.Tags, error)
	TagsGetTags(ctx context.Context, tenant string) ([]models.Tags, int64, error)
	TagsPushTag(ctx context.Context, tagName, tenantID string) error
	TagsBulkDeleteTag(ctx context.Context, tenant, tagName string) (int64, error)
}
