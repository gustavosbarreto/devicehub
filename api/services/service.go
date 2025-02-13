package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type APIService struct {
	*service
}

var _ Service = (*APIService)(nil)

type service struct {
	store     store.Store
	privKey   *rsa.PrivateKey
	pubKey    *rsa.PublicKey
	cache     cache.Cache
	client    internalclient.Client
	locator   geoip.Locator
	validator *validator.Validator
}

//go:generate mockery --name Service --filename services.go
type Service interface {
	BillingInterface
	TagsService
	DeviceService
	UserService
	SSHKeysService
	SessionService
	NamespaceService
	MemberService
	AuthService
	StatsService
	SetupService
	SystemService
	APIKeyService
}

type Option func(service *APIService)

func WithLocator(locator geoip.Locator) Option {
	return func(service *APIService) {
		service.locator = locator
	}
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, c internalclient.Client, options ...Option) *APIService {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	service := &APIService{
		service: &service{
			store,
			privKey,
			pubKey,
			cache,
			c,
			geoip.NewNullGeoLite(),
			validator.New(),
		},
	}

	for _, option := range options {
		option(service)
	}

	return service
}
