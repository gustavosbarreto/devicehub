package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error
	DeviceRenameTag(ctx context.Context, tenant, oldTag, newTag string) error
	DeviceDeleteTag(ctx context.Context, tenant, tag string) error
}
