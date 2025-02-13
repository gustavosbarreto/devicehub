package store

//go:generate mockery --name Store --filename store.go
type Store interface {
	TagsStore
	DeviceStore
	SessionStore
	UserStore
	NamespaceStore
	PublicKeyStore
	PrivateKeyStore
	StatsStore
	APIKeyStore
	TransactionStore
	SystemStore

	Options() QueryOptions
}
