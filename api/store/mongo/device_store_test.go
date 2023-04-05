package mongo

import (
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)
}

func TestDeviceGet(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGet(data.Context, models.UID(data.Device.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceRename(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceRename(data.Context, models.UID(data.Device.UID), "newHostname")
	assert.NoError(t, err)
}

func TestDeviceLookup(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateStatus(data.Context, models.UID(data.Device.UID), "accepted")
	assert.NoError(t, err)

	d, err := mongostore.DeviceLookup(data.Context, data.Namespace.Name, "hostname")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceUpdateStatus(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateStatus(data.Context, models.UID(data.Device.UID), "accepted")
	assert.NoError(t, err)
}

func TestDeviceSetOnline(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceSetOnline(data.Context, models.UID(data.Device.UID), true)
	assert.NoError(t, err)
}

func TestDeviceUpdateOnline(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateOnline(data.Context, models.UID(data.Device.UID), true)
	assert.NoError(t, err)
}

func TestDeviceUpdateLastSeen(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateLastSeen(data.Context, models.UID(data.Device.UID), time.Now())
	assert.NoError(t, err)
}

func TestDeviceGetByMac(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByMac(data.Context, "mac", "00000000-0000-4000-0000-000000000000", "pending")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceGetByName(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByName(data.Context, "hostname", "00000000-0000-4000-0000-000000000000")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceGetByUID(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(data.Context, models.UID(data.Device.UID), "00000000-0000-4000-0000-000000000000")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDevicesList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	devices, count, err := mongostore.DeviceList(data.Context, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc", store.DeviceListModeDefault)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, devices)
}

func TestDeviceListByUsage(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	sessions := make([]models.Session, 0)

	quantities := []int{10, 5, 3, 1, 1, 0, 0, 0}

	for i, q := range quantities {
		for j := 0; j < q; j++ {
			sessions = append(sessions, models.Session{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				DeviceUID: models.UID(fmt.Sprintf("%s%d", "uid", i+1)),
			})
		}
	}

	sessionsInterfaces := make([]interface{}, len(sessions))

	for i, v := range sessions {
		sessionsInterfaces[i] = v
	}
	_, err = db.Client().Database("test").Collection("sessions").InsertMany(data.Context, sessionsInterfaces)
	assert.NoError(t, err)

	devices, err := mongostore.DeviceListByUsage(data.Context, data.Namespace.TenantID)
	expectedUIDs := []models.UID{"uid1", "uid2", "uid3"}

	assert.NoError(t, err)
	assert.Equal(t, len(expectedUIDs), len(devices))

	for i, device := range devices {
		assert.Equal(t, expectedUIDs[i], device)
	}
}

func TestDeviceChooser(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	devices := make([]models.Device, 0)

	devicesInterfaces := make([]interface{}, 5)

	for i := 0; i < 5; i++ {
		devices = append(devices, models.Device{
			UID:      fmt.Sprintf("%s%d", "uid", i+1),
			TenantID: "00000000-0000-4000-0000-000000000000",
			Status:   "accepted",
		})
	}

	for i, v := range devices {
		devicesInterfaces[i] = v
	}

	_, err = db.Client().Database("test").Collection("devices").InsertMany(data.Context, devicesInterfaces)
	assert.NoError(t, err)

	err = mongostore.DeviceChooser(data.Context, data.Namespace.TenantID, []string{"uid1", "uid2", "uid5"})
	assert.NoError(t, err)

	devices, _, err = mongostore.DeviceList(data.Context, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc", store.DeviceListModeDefault)
	assert.NoError(t, err)

	pending := make([]string, 0)

	expected := []string{"uid3", "uid4"}

	for _, dev := range devices {
		if dev.Status == "pending" {
			pending = append(pending, dev.UID)
		}
	}

	assert.Equal(t, expected, pending)
}
