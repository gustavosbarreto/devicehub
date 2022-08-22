package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/geoip"
)

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
	cache   cache.Cache
	client  interface{}
	locator geoip.Locator
}

type Service interface {
	TagsService
	DeviceService
	DeviceTags
	UserService
	SSHKeysService
	SSHKeysTagsService
	SessionService
	NamespaceService
	AuthService
	StatsService
	SetupService
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, c interface{}, l geoip.Locator) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey, cache, c, l}
}

// FIXME: private function.
//
// Cloud needs to be able to access this function to convert some status to a API ErrReport.
func HandleStatusResponse(status int) error {
	if status == 200 || status == 402 || status == 400 {
		return nil
	}

	return ErrReport
}
