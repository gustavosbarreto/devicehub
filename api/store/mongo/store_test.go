package mongo_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	log "github.com/sirupsen/logrus"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

var (
	srv = &dbtest.Server{}
	db  *mongodb.Database
	s   store.Store
)

const (
	fixtureAPIKeys          = "api-key"           // Check "store.mongo.fixtures.api-keys" for fixture info
	fixtureConnectedDevices = "connected_devices" // Check "store.mongo.fixtures.connected_devices" for fixture info
	fixtureDevices          = "devices"           // Check "store.mongo.fixtures.devices" for fixture info
	fixtureSessions         = "sessions"          // Check "store.mongo.fixtures.sessions" for fixture info
	fixtureActiveSessions   = "active_sessions"   // Check "store.mongo.fixtures.active_sessions" for fixture info
	fixtureFirewallRules    = "firewall_rules"    // Check "store.mongo.fixtures.firewall_rules" for fixture info
	fixturePublicKeys       = "public_keys"       // Check "store.mongo.fixtures.public_keys" for fixture info
	fixturePrivateKeys      = "private_keys"      // Check "store.mongo.fixtures.private_keys" for fixture info
	fixtureUsers            = "users"             // Check "store.mongo.fixtures.users" for fixture iefo
	fixtureNamespaces       = "namespaces"        // Check "store.mongo.fixtures.namespaces" for fixture info
	fixtureRecoveryTokens   = "recovery_tokens"   // Check "store.mongo.fixtures.recovery_tokens" for fixture info
	fixtureTags             = "tags"              // Check "store.mongo.fixtures.tags" for fixture info
)

func TestMain(m *testing.M) {
	log.Info("Starting store tests")

	ctx := context.Background()

	srv.Container.Database = "test"
	_, file, _, _ := runtime.Caller(0)
	srv.Fixtures.Root = filepath.Join(filepath.Dir(file), "fixtures")
	srv.Fixtures.PreInsertFuncs = []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("users", "_id"),
		mongotest.SimpleConvertTime("users", "created_at"),
		mongotest.SimpleConvertTime("users", "last_login"),
		mongotest.SimpleConvertObjID("public_keys", "_id"),
		mongotest.SimpleConvertBytes("public_keys", "data"),
		mongotest.SimpleConvertTime("public_keys", "created_at"),
		mongotest.SimpleConvertObjID("private_keys", "_id"),
		mongotest.SimpleConvertBytes("private_keys", "data"),
		mongotest.SimpleConvertTime("private_keys", "created_at"),
		mongotest.SimpleConvertObjID("namespaces", "_id"),
		mongotest.SimpleConvertTime("namespaces", "created_at"),
		mongotest.SimpleConvertObjID("devices", "_id"),
		mongotest.SimpleConvertTime("devices", "created_at"),
		mongotest.SimpleConvertTime("devices", "last_seen"),
		mongotest.SimpleConvertTime("devices", "status_updated_at"),
		mongotest.SimpleConvertTime("connected_devices", "last_seen"),
		mongotest.SimpleConvertObjID("firewall_rules", "_id"),
		mongotest.SimpleConvertObjID("sessions", "_id"),
		mongotest.SimpleConvertTime("sessions", "started_at"),
		mongotest.SimpleConvertTime("sessions", "last_seen"),
		mongotest.SimpleConvertObjID("active_sessions", "_id"),
		mongotest.SimpleConvertTime("active_sessions", "last_seen"),
		mongotest.SimpleConvertObjID("tags", "_id"),
	}

	if err := srv.Up(ctx); err != nil {
		log.WithError(err).Error("Failed to UP the mongodb container")
		os.Exit(1)
	}

	log.Info("Connecting to ", srv.Container.ConnectionString)

	var err error

	s, err = mongo.NewStore(ctx, srv.Container.ConnectionString+"/"+srv.Container.Database, cache.NewNullCache())
	if err != nil {
		log.WithError(err).Error("Failed to create the mongodb store")
		os.Exit(1)
	}

	store := s.(*mongo.Store)
	db = store.GetDB()

	code := m.Run()

	log.Info("Stopping store tests")
	if err := srv.Down(ctx); err != nil {
		log.WithError(err).Error("Failed to DOWN the mongodb container")
		os.Exit(1)
	}

	os.Exit(code)
}
