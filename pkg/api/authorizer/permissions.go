package authorizer

type Permission int

const (
	DeviceAccept Permission = iota
	DeviceReject
	DeviceUpdate
	DeviceRemove
	DeviceConnect
	DeviceRename
	DeviceDetails

	TagCreate
	TagUpdate
	TagDelete

	SessionPlay
	SessionClose
	SessionRemove
	SessionDetails

	FirewallCreate
	FirewallEdit
	FirewallRemove
	FirewallAddTag
	FirewallRemoveTag
	FirewallUpdateTag

	PublicKeyCreate
	PublicKeyEdit
	PublicKeyRemove
	PublicKeyAddTag
	PublicKeyRemoveTag
	PublicKeyUpdateTag

	NamespaceUpdate
	NamespaceAddMember
	NamespaceRemoveMember
	NamespaceEditMember
	NamespaceEnableSessionRecord
	NamespaceDelete

	BillingCreateCustomer
	BillingChooseDevices
	BillingAddPaymentMethod
	BillingUpdatePaymentMethod
	BillingRemovePaymentMethod
	BillingCancelSubscription
	BillingCreateSubscription
	BillingGetPaymentMethod
	BillingGetSubscription

	APIKeyCreate
	APIKeyUpdate
	APIKeyDelete

	ConnectorDelete
	ConnectorUpdate
	ConnectorSet

	TunnelsCreate
	TunnelsDelete
)

var observerPermissions = []Permission{
	DeviceConnect,
	DeviceDetails,

	SessionDetails,
}

var operatorPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionDetails,
}

var adminPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,
	FirewallAddTag,
	FirewallRemoveTag,
	FirewallUpdateTag,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,
	PublicKeyAddTag,
	PublicKeyRemoveTag,
	PublicKeyUpdateTag,

	NamespaceUpdate,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,

	APIKeyCreate,
	APIKeyUpdate,
	APIKeyDelete,

	ConnectorDelete,
	ConnectorUpdate,
	ConnectorSet,

	TunnelsCreate,
	TunnelsDelete,
}

var ownerPermissions = []Permission{
	DeviceAccept,
	DeviceReject,
	DeviceRemove,
	DeviceConnect,
	DeviceRename,
	DeviceDetails,
	DeviceUpdate,

	TagCreate,
	TagUpdate,
	TagDelete,

	SessionPlay,
	SessionClose,
	SessionRemove,
	SessionDetails,

	FirewallCreate,
	FirewallEdit,
	FirewallRemove,
	FirewallAddTag,
	FirewallRemoveTag,
	FirewallUpdateTag,

	PublicKeyCreate,
	PublicKeyEdit,
	PublicKeyRemove,
	PublicKeyAddTag,
	PublicKeyRemoveTag,
	PublicKeyUpdateTag,

	NamespaceUpdate,
	NamespaceAddMember,
	NamespaceRemoveMember,
	NamespaceEditMember,
	NamespaceEnableSessionRecord,
	NamespaceDelete,

	BillingCreateCustomer,
	BillingChooseDevices,
	BillingAddPaymentMethod,
	BillingUpdatePaymentMethod,
	BillingRemovePaymentMethod,
	BillingCancelSubscription,
	BillingCreateSubscription,
	BillingGetSubscription,

	APIKeyCreate,
	APIKeyUpdate,
	APIKeyDelete,

	ConnectorDelete,
	ConnectorUpdate,
	ConnectorSet,

	TunnelsCreate,
	TunnelsDelete,
}
