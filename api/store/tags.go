package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type TagsStore interface {
	// TagCreate creates new tag.
	//
	// It returns the inserted ID or an error if any.
	TagCreate(ctx context.Context, tag *models.Tag) (insertedID string, err error)

	// TagConflicts checks for uniqueness violations of tag attributes within a namespace.
	// Only non-zero values in the target are checked for conflicts.
	//
	// Example:
	//     ctx := context.Background()
	//     conflicts, has, err := store.TagConflicts(ctx, "tenant123", &models.TagConflicts{Name: "development"})
	//     println(conflicts) // => []string{"name"}
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) (conflicts []string, has bool, err error)

	// TagList retrieves a list of tags based on the provided filters and pagination settings. When tenantID is
	// empty, it returns all tags.
	//
	// It returns the list of tags, the total count of matching documents (ignoring pagination), and
	// an error if any.
	TagList(ctx context.Context, tenantID string, paginator query.Paginator, filters query.Filters, sorter query.Sorter) (tags []models.Tag, totalCount int, err error)

	// TagGetByID retrieves a tag identified by the given ID.
	//
	// It returns the tag or an error if any.
	TagGetByID(ctx context.Context, id string) (tag *models.Tag, err error)

	// TagGetByName retrieves a tag identified by the given name within a namespace with the given tenant ID.
	//
	// It returns the tag or an error if any.
	TagGetByName(ctx context.Context, tenantID, name string) (tag *models.Tag, err error)

	// TagUpdate updates a tag identified by the given name within a namespace with the given tenant ID.
	//
	// It returns an error, if any, or store.ErrNoDocuments if the tag does not exist.
	TagUpdate(ctx context.Context, tenantID, name string, changes *models.TagChanges) (err error)

	// TagPushToTarget pushs an existent tag to the provided target.
	//
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	TagPushToTarget(ctx context.Context, tenantID, name string, target models.TagTarget, targetID string) (err error)

	// TagPullFromTarget removes a tag from tagged documents in a namespace. If targetsID is empty it removes the tag from
	// all documents of the selected target type. If targetsID contains specific target IDs it only removes the tag from those
	// documents.
	//
	// Returns ErrNoDocuments if no matching documents found or other errors from the operation.
	TagPullFromTarget(ctx context.Context, tenantID, name string, target models.TagTarget, targetsID ...string) (err error)

	// TagUpdate delete a tag identified by the given name within a namespace with the given tenant ID.
	//
	// It returns an error, if any, or store.ErrNoDocuments if the tag does not exist.
	TagDelete(ctx context.Context, tenantID, name string) (err error)
}
