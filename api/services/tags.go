package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type TagsService interface {
	// CreateTag creates a new tag in the specified tenant namespace.
	//
	// Tags can share the same attributes (e.g. name) if they belong to different tenants.
	// For example, tenant1 and tenant2 can each have a tag named "production".
	//
	// It returns the insertedID, an array of conflicting field names, e.g. `["name"]` and an error if any.
	CreateTag(ctx context.Context, req *requests.CreateTag) (insertedID string, conflicts []string, err error)

	// PushTagTo adds an existing tag in a namespace to a target document (e.g. Device, PublicKey, FirewallRule).
	//
	// Returns ErrNamespaceNotFound if namespace not found, ErrTagNotFound if tag not found, or other errors if operation fails.
	PushTagTo(ctx context.Context, target models.TagTarget, req *requests.PushTag) (err error)

	// PullTagFrom removes a tag from a target document in a namespace. The tag itself is not deleted.
	// If no other documents reference the tag, it becomes orphaned but remains available for future use.
	//
	// Returns ErrNamespaceNotFound if namespace not found, ErrTagNotFound if tag not found, or other errors if operation fails.
	PullTagFrom(ctx context.Context, target models.TagTarget, req *requests.PullTag) (err error)

	// ListTags retrieves a batch of tags that belong to the given namespace.
	//
	// It returns the list of tags with pagination, an integer representing the total count of tags in the
	// database, ignoring pagination, and an error if any.
	ListTags(ctx context.Context, req *requests.ListTags) (tags []models.Tag, totalCount int, err error)

	// UpdateTag updates a tag with the specified name in the specified namespace.
	//
	// It returns an array of conflicting field names, e.g. `["name"]` and an error if any.
	UpdateTag(ctx context.Context, req *requests.UpdateTag) (conflicts []string, err error)

	// DeleteTag deletes a tag with the specified name in the specified namespace.
	//
	// It returns an error if any.
	DeleteTag(ctx context.Context, req *requests.DeleteTag) (err error)
}

func (s *service) CreateTag(ctx context.Context, req *requests.CreateTag) (string, []string, error) {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return "", []string{}, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if conflicts, has, err := s.store.TagConflicts(ctx, req.TenantID, &models.TagConflicts{Name: req.Name}); has || err != nil {
		return "", conflicts, err
	}

	insertedID, err := s.store.TagCreate(ctx, &models.Tag{Name: req.Name, TenantID: req.TenantID})
	if err != nil {
		return "", []string{}, err
	}

	return insertedID, []string{}, nil
}

func (s *service) PushTagTo(ctx context.Context, target models.TagTarget, req *requests.PushTag) (err error) {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, err := s.store.TagGetByName(ctx, req.TenantID, req.Name); err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	return s.store.TagPushToTarget(ctx, req.TenantID, req.Name, target, req.TargetID)
}

func (s *service) PullTagFrom(ctx context.Context, target models.TagTarget, req *requests.PullTag) (err error) {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, err := s.store.TagGetByName(ctx, req.TenantID, req.Name); err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	return s.store.TagPullFromTarget(ctx, req.TenantID, req.Name, target, req.TargetID)
}

func (s *service) ListTags(ctx context.Context, req *requests.ListTags) ([]models.Tag, int, error) {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return []models.Tag{}, 0, NewErrNamespaceNotFound(req.TenantID, err)
	}

	tags, totalCount, err := s.store.TagList(ctx, req.TenantID, req.Paginator, req.Filters, req.Sorter)
	if err != nil {
		return []models.Tag{}, 0, err
	}

	return tags, totalCount, nil
}

func (s *service) UpdateTag(ctx context.Context, req *requests.UpdateTag) ([]string, error) {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return []string{}, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, err := s.store.TagGetByName(ctx, req.TenantID, req.Name); err != nil {
		return []string{}, NewErrTagNotFound(req.Name, err)
	}

	conflictsAttrs := &models.TagConflicts{}
	if req.NewName != "" && req.NewName != req.Name {
		conflictsAttrs.Name = req.NewName
	}

	if conflicts, has, err := s.store.TagConflicts(ctx, req.TenantID, conflictsAttrs); has || err != nil {
		return conflicts, err
	}

	if err := s.store.TagUpdate(ctx, req.TenantID, req.Name, &models.TagChanges{Name: req.NewName}); err != nil {
		return nil, err
	}

	return []string{}, nil
}

func (s *service) DeleteTag(ctx context.Context, req *requests.DeleteTag) error {
	if _, err := s.store.NamespaceGet(ctx, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, err := s.store.TagGetByName(ctx, req.TenantID, req.Name); err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	return s.store.WithTransaction(ctx, s.deleteTagCallback(req))
}

func (s *service) deleteTagCallback(req *requests.DeleteTag) store.TransactionCb {
	return func(ctx context.Context) error {
		for _, target := range models.TagTargets() {
			if err := s.store.TagPullFromTarget(ctx, req.TenantID, req.Name, target); err != nil {
				return err
			}
		}

		return s.store.TagDelete(ctx, req.TenantID, req.Name)
	}
}
