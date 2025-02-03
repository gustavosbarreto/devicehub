package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceAcceptable uint

const (
	// DeviceAcceptableIfNotAccepted is used to indicate the all devices not accepted will be defined as "acceptabled".
	DeviceAcceptableIfNotAccepted DeviceAcceptable = iota + 1
	// DeviceAcceptableFromRemoved is used to indicate that the namepsace's device maxium number of devices has been
	// reached and should set the "acceptable" value to true for devices that were recently removed.
	DeviceAcceptableFromRemoved
	// DeviceAcceptableAsFalse set acceptable to false to all returned devices.
	DeviceAcceptableAsFalse
)

type DeviceStore interface {
	// DeviceList retrieves a list of devices based on the provided filters and pagination settings. A list of options can be
	// passed to inject additional data into each device in the list.
	//
	// It returns the list of namespaces, the total count of matching documents (ignoring pagination), and
	// an error if any.
	DeviceList(ctx context.Context, status models.DeviceStatus, pagination query.Paginator, filters query.Filters, sorter query.Sorter, acceptable DeviceAcceptable, opts ...DeviceQueryOption) ([]models.Device, int, error)

	DeviceGet(ctx context.Context, uid models.UID, opts ...DeviceQueryOption) (*models.Device, error)
	DeviceUpdate(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error
	DeviceDelete(ctx context.Context, uid models.UID) error
	DeviceCreate(ctx context.Context, d models.Device, hostname string) error
	DeviceRename(ctx context.Context, uid models.UID, hostname string) error
	DeviceLookup(ctx context.Context, namespace, hostname string, opts ...DeviceQueryOption) (*models.Device, error)
	DeviceUpdateOnline(ctx context.Context, uid models.UID, online bool) error
	DeviceUpdateLastSeen(ctx context.Context, uid models.UID, ts time.Time) error
	DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error
	DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus, opts ...DeviceQueryOption) (*models.Device, error)
	DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus, opts ...DeviceQueryOption) (*models.Device, error)
	DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string, opts ...DeviceQueryOption) (*models.Device, error)
	DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error
	DeviceListByUsage(ctx context.Context, tenantID string) ([]models.UID, error)
	DeviceChooser(ctx context.Context, tenantID string, chosen []string) error
	DeviceRemovedCount(ctx context.Context, tenant string) (int64, error)
	DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error)
	DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error
	DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error
	DeviceRemovedList(ctx context.Context, tenant string, pagination query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.DeviceRemoved, int, error)
	DeviceCreatePublicURLAddress(ctx context.Context, uid models.UID) error
	DeviceGetByPublicURLAddress(ctx context.Context, address string, opts ...DeviceQueryOption) (*models.Device, error)

	// DeviceSetOnline receives a list of devices to mark as online. For each device in the array, it will upsert
	// a connected device entry; each UID must exists in the "devices" collection.
	DeviceSetOnline(ctx context.Context, connectedDevices []models.ConnectedDevice) error

	// DeviceSetOffline sets a device's status to offline using its UID.
	DeviceSetOffline(ctx context.Context, uid string) error
}
