package services

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for service's error.
const ErrLayer = "service"

const (
	// ErrCodeNotFound is the error code for when a resource is not found.
	ErrCodeNotFound = iota + 1
	// ErrCodeDuplicated is the error code for when a resource is duplicated.
	ErrCodeDuplicated
	// ErrCodeLimit is the error code for when a resource is reached the limit.
	ErrCodeLimit
	// ErrCodeInvalid is the error code for when a resource is invalid.
	ErrCodeInvalid
	// ErrCodePayment is the error code for when a resource required payment.
	ErrCodePayment
	// ErrCodeUnauthorized is the error code for when the access to resource is unauthorized.
	ErrCodeUnauthorized
	// ErrCodeForbidden is the error code for when the access to resource is forbidden.
	ErrCodeForbidden
	// ErrCodeStore is the error code for when the store function fails. The store function is responsible for execute
	// the main service action.
	ErrCodeStore
	// ErrCodeNoContentChange is the error that occurs when the store function does not change any resource. Generally used in
	// update methods.
	ErrCodeNoContentChange
	// ErrCodeCreated is the error code to be used when the resource was created, but the following operations failed.
	ErrCodeCreated
)

// ErrDataNotFound structure should be used to add errors.Data to an error when the resource is not found.
type ErrDataNotFound struct {
	// ID is the identifier of the resource.
	ID string
}

// ErrDataDuplicated structure should be used to add errors.Data to an error when the resource is duplicated.
type ErrDataDuplicated struct {
	// Values is used to identify the duplicated resource.
	Values []string
}

// ErrDataLimit structure should be used to add errors.Data to an error when the resource is reached the limit.
type ErrDataLimit struct {
	// Limit is the max number of resources.
	Limit int
}

// ErrDataInvalid structure should be used to add errors.Data to an error when the resource is invalid.
type ErrDataInvalid struct {
	// Data is a key-value map of the invalid fields. key must be the field name what is invalid and value must be the
	// value of the field.
	Data map[string]interface{}
}

var (
	ErrReport                       = errors.New("report error", ErrLayer, ErrCodeInvalid)
	ErrPaymentRequired              = errors.New("payment required", ErrLayer, ErrCodePayment)
	ErrEvaluate                     = errors.New("evaluate error", ErrLayer, ErrCodeInvalid)
	ErrNoContentChange              = errors.New("no content change", ErrLayer, ErrCodeNoContentChange)
	ErrNotFound                     = errors.New("not found", ErrLayer, ErrCodeNotFound)
	ErrBadRequest                   = errors.New("bad request", ErrLayer, ErrCodeInvalid)
	ErrUnauthorized                 = errors.New("unauthorized", ErrLayer, ErrCodeInvalid)
	ErrForbidden                    = errors.New("forbidden", ErrLayer, ErrCodeNotFound)
	ErrUserNotFound                 = errors.New("user not found", ErrLayer, ErrCodeNotFound)
	ErrUserInvalid                  = errors.New("user invalid", ErrLayer, ErrCodeInvalid)
	ErrUserDuplicated               = errors.New("user duplicated", ErrLayer, ErrCodeDuplicated)
	ErrUserPasswordInvalid          = errors.New("user password invalid", ErrLayer, ErrCodeInvalid)
	ErrUserPasswordDuplicated       = errors.New("user password is equal to new password", ErrLayer, ErrCodeDuplicated)
	ErrUserPasswordNotMatch         = errors.New("user password does not match to the current password", ErrLayer, ErrCodeInvalid)
	ErrUserNotConfirmed             = errors.New("user not confirmed", ErrLayer, ErrCodeForbidden)
	ErrUserUpdate                   = errors.New("user update", ErrLayer, ErrCodeStore)
	ErrNamespaceNotFound            = errors.New("namespace not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceInvalid             = errors.New("namespace invalid", ErrLayer, ErrCodeInvalid)
	ErrNamespaceList                = errors.New("namespace member list", ErrLayer, ErrCodeNotFound)
	ErrNamespaceDuplicated          = errors.New("namespace duplicated", ErrLayer, ErrCodeDuplicated)
	ErrNamespaceMemberNotFound      = errors.New("member not found", ErrLayer, ErrCodeNotFound)
	ErrNamespaceMemberInvalid       = errors.New("member invalid", ErrLayer, ErrCodeInvalid)
	ErrNamespaceMemberFillData      = errors.New("member fill data", ErrLayer, ErrCodeInvalid)
	ErrNamespaceMemberDuplicated    = errors.New("member duplicated", ErrLayer, ErrCodeDuplicated)
	ErrNamespaceCreateStore         = errors.New("namespace create store", ErrLayer, ErrCodeStore)
	ErrMaxTagReached                = errors.New("tag limit reached", ErrLayer, ErrCodeLimit)
	ErrDuplicateTagName             = errors.New("tag duplicated", ErrLayer, ErrCodeDuplicated)
	ErrTagNameNotFound              = errors.New("tag not found", ErrLayer, ErrCodeNotFound)
	ErrTagInvalid                   = errors.New("tag invalid", ErrLayer, ErrCodeInvalid)
	ErrNoTags                       = errors.New("no tags has found", ErrLayer, ErrCodeNotFound)
	ErrConflictName                 = errors.New("name duplicated", ErrLayer, ErrCodeDuplicated)
	ErrInvalidFormat                = errors.New("invalid format", ErrLayer, ErrCodeInvalid)
	ErrDeviceNotFound               = errors.New("device not found", ErrLayer, ErrCodeNotFound)
	ErrDeviceInvalid                = errors.New("device invalid", ErrLayer, ErrCodeInvalid)
	ErrDeviceDuplicated             = errors.New("device duplicated", ErrLayer, ErrCodeDuplicated)
	ErrDeviceLookupNotFound         = errors.New("device lookup not found", ErrLayer, ErrCodeNotFound)
	ErrDeviceLimit                  = errors.New("device limit reached", ErrLayer, ErrCodePayment)
	ErrDeviceStatusInvalid          = errors.New("device status invalid", ErrLayer, ErrCodeInvalid)
	ErrDeviceStatusAccepted         = errors.New("device status accepted", ErrLayer, ErrCodeInvalid)
	ErrDeviceCreate                 = errors.New("device create", ErrLayer, ErrCodeStore)
	ErrDeviceSetOnline              = errors.New("device set online", ErrLayer, ErrCodeStore)
	ErrMaxDeviceCountReached        = errors.New("maximum number of accepted devices reached", ErrLayer, ErrCodeLimit)
	ErrDuplicatedDeviceName         = errors.New("device name duplicated", ErrLayer, ErrCodeDuplicated)
	ErrPublicKeyDuplicated          = errors.New("public key duplicated", ErrLayer, ErrCodeDuplicated)
	ErrPublicKeyNotFound            = errors.New("public key not found", ErrLayer, ErrCodeNotFound)
	ErrPublicKeyInvalid             = errors.New("public key invalid", ErrLayer, ErrCodeInvalid)
	ErrPublicKeyNoTags              = errors.New("public key has no tags", ErrLayer, ErrCodeInvalid)
	ErrPublicKeyDataInvalid         = errors.New("public key data invalid", ErrLayer, ErrCodeInvalid)
	ErrPublicKeyFilter              = errors.New("public key cannot have more than one filter at same time", ErrLayer, ErrCodeInvalid)
	ErrTokenSigned                  = errors.New("token signed", ErrLayer, ErrCodeInvalid)
	ErrTypeAssertion                = errors.New("type assertion failed", ErrLayer, ErrCodeInvalid)
	ErrSessionNotFound              = errors.New("session not found", ErrLayer, ErrCodeNotFound)
	ErrAuthInvalid                  = errors.New("auth invalid", ErrLayer, ErrCodeInvalid)
	ErrAuthUnathorized              = errors.New("auth unauthorized", ErrLayer, ErrCodeUnauthorized)
	ErrNamespaceLimitReached        = errors.New("namespace limit reached", ErrLayer, ErrCodeLimit)
	ErrDeviceRemovedCount           = errors.New("device removed count", ErrLayer, ErrCodeNotFound)
	ErrDeviceRemovedInsert          = errors.New("device removed insert", ErrLayer, ErrCodeStore)
	ErrDeviceRemovedFull            = errors.New("device removed full", ErrLayer, ErrCodePayment)
	ErrDeviceRemovedDelete          = errors.New("device removed delete", ErrLayer, ErrCodeStore)
	ErrDeviceRemovedGet             = errors.New("device removed get", ErrLayer, ErrCodeNotFound)
	ErrBillingReportNamespaceDelete = errors.New("billing report namespace delete", ErrLayer, ErrCodePayment)
	ErrBillingReportDevice          = errors.New("billing report device", ErrLayer, ErrCodePayment)
	ErrBillingEvaluate              = errors.New("billing evaluate", ErrLayer, ErrCodePayment)
	ErrSameTags                     = errors.New("trying to update tags with the same content", ErrLayer, ErrCodeNoContentChange)
	ErrAPIKeyNotFound               = errors.New("APIKey not found", ErrLayer, ErrCodeNotFound)
	ErrAPIKeyDuplicated             = errors.New("APIKey duplicated", ErrLayer, ErrCodeDuplicated)
	ErrAuthForbidden                = errors.New("user is authenticated but cannot access this resource", ErrLayer, ErrCodeForbidden)
	ErrRoleInvalid                  = errors.New("role is invalid", ErrLayer, ErrCodeForbidden)
	ErrNamespaceIPInvalid           = errors.New("ip is invalid", ErrLayer, ErrCodeForbidden)
	ErrNamespaceIPNotPrivate        = errors.New("ip is not a private address", ErrLayer, ErrCodeForbidden)
)

func NewErrRoleInvalid() error {
	return ErrRoleInvalid
}

// NewErrNotFound returns an error with the ErrDataNotFound and wrap an error.
func NewErrNoContentChange(err error, next error) error {
	return errors.Wrap(err, next)
}

// NewErrNotFound returns an error with the ErrDataNotFound and wrap an error.
func NewErrNotFound(err error, id string, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataNotFound{ID: id}), next)
}

// NewErrInvalid returns an error with the ErrDataInvalid and wrap an error.
func NewErrInvalid(err error, data map[string]interface{}, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataInvalid{Data: data}), next)
}

// NewErrDuplicated returns an error with the ErrDataDuplicated and wrap an error.
func NewErrDuplicated(err error, values []string, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataDuplicated{Values: values}), next)
}

// NewErrLimit returns an error with the ErrDataLimit and wrap an error.
func NewErrLimit(err error, limit int, next error) error {
	return errors.Wrap(errors.WithData(err, ErrDataLimit{Limit: limit}), next)
}

// NewErrStore return an error to be used when the main store function fails.
//
// A service can make n calls to store's function, but each service has your main action; what it was made to do. For
// this case, to use when the main store's function fails, this error was intended to be used.
func NewErrStore(err error, data interface{}, next error) error {
	return errors.Wrap(errors.WithData(err, data), next)
}

// NewErrUnathorized returns a error to be used when the access to a resource is not authorized.
func NewErrUnathorized(err error, next error) error {
	return errors.Wrap(err, next)
}

// NewErrBadRequest returns a error to be used when the access to a resource is not authorized.
func NewErrRequest(err error, next error) error {
	return errors.Wrap(err, next)
}

// NewErrForbidden return a error to be used when the access to a resource is forbidden.
func NewErrForbidden(err error, next error) error {
	return errors.Wrap(err, next)
}

// NewErrNamespaceNotFound returns an error when the namespace is not found.
func NewErrNamespaceNotFound(id string, next error) error {
	return NewErrNotFound(ErrNamespaceNotFound, id, next)
}

// NewErrAPIKeyNotFound returns an error when the APIKey is not found.
func NewErrAPIKeyNotFound(name string, next error) error {
	return NewErrNotFound(ErrAPIKeyNotFound, name, next)
}

func NewErrAPIKeyInvalid(name string) error {
	return NewErrAuthInvalid(map[string]interface{}{"api-key": name}, nil)
}

// NewErrAPIKeyDuplicated returns an error when the APIKey name is duplicated.
func NewErrAPIKeyDuplicated(conflicts []string) error {
	return NewErrDuplicated(ErrAPIKeyDuplicated, conflicts, nil)
}

// NewErrTagInvalid returns an error when the tag is invalid.
func NewErrTagInvalid(tag string, next error) error {
	return NewErrInvalid(ErrTagInvalid, map[string]interface{}{"name": tag}, next)
}

// NewErrSameTags returns an error when the
func NewErrSameTags() error {
	return NewErrNoContentChange(ErrSameTags, nil)
}

// NewErrTagEmpty returns an error when the none tag is found.
func NewErrTagEmpty(tenant string, next error) error {
	return NewErrNotFound(ErrNoTags, tenant, next)
}

// NewErrTagNotFound returns an error when the tag is not found.
func NewErrTagNotFound(tag string, next error) error {
	return NewErrNotFound(ErrTagNameNotFound, tag, next)
}

// NewErrTagDuplicated returns an error when the tag is duplicated.
func NewErrTagDuplicated(tag string, next error) error {
	return NewErrDuplicated(ErrDuplicateTagName, []string{tag}, next)
}

// NewErrUserNotFound returns an error when the user is not found.
func NewErrUserNotFound(id string, next error) error {
	return NewErrNotFound(ErrUserNotFound, id, next)
}

// NewErrUserInvalid returns an error when the user is invalid.
func NewErrUserInvalid(data map[string]interface{}, next error) error {
	return NewErrInvalid(ErrUserInvalid, data, next)
}

// NewErrUserDuplicated returns an error when the user is duplicated.
func NewErrUserDuplicated(values []string, next error) error {
	return NewErrDuplicated(ErrUserDuplicated, values, next)
}

// NewErrUserPasswordInvalid returns an error when the user's password is invalid.
func NewErrUserPasswordInvalid(next error) error {
	return NewErrInvalid(ErrUserPasswordInvalid, nil, next)
}

// NewErrUserPasswordDuplicated returns an error when the user's current password is equal to new password.
func NewErrUserPasswordDuplicated(next error) error {
	return NewErrDuplicated(ErrUserPasswordDuplicated, nil, next)
}

// NewErrUserPasswordNotMatch returns an error when the user's password doesn't match with the current password.
func NewErrUserPasswordNotMatch(next error) error {
	return NewErrInvalid(ErrUserPasswordNotMatch, nil, next)
}

// NewErrPublicKeyNotFound returns an error when the public key is not found.
func NewErrPublicKeyNotFound(id string, next error) error {
	return NewErrNotFound(ErrPublicKeyNotFound, id, next)
}

// NewErrPublicKeyInvalid returns an error when the public key is invalid.
func NewErrPublicKeyInvalid(data map[string]interface{}, next error) error {
	return NewErrInvalid(ErrPublicKeyInvalid, data, next)
}

// NewErrTagLimit returns an error when the tag limit is reached.
func NewErrTagLimit(limit int, next error) error {
	return NewErrLimit(ErrMaxTagReached, limit, next)
}

// NewErrPublicKeyDuplicated returns an error when the public key is duplicated.
func NewErrPublicKeyDuplicated(values []string, next error) error {
	return NewErrDuplicated(ErrPublicKeyDuplicated, values, next)
}

// NewErrPublicKeyTagsEmpty returns an error when the public key has no tags.
func NewErrPublicKeyTagsEmpty(next error) error {
	return NewErrNotFound(ErrPublicKeyNoTags, "", next)
}

// NewErrPublicKeyDataInvalid returns an error when the public key data is invalid.
func NewErrPublicKeyDataInvalid(value []byte, next error) error {
	// FIXME: literal assignment.
	//
	// The literal assignment of value to a map's key "Data" is required because the service doesn't have conscious about
	// the models.PublicKey field when it validate the public key data. When validating other fields, the validation
	// function return the field and value what is invalid, but in this case, the validation occur by the check of
	// ssh.ParseAuthorizedKey result.
	//
	// To fix this, I believe that all extra validation could be set as structure methods, centralizing the structure
	// value agreement.
	//
	// For now, there are a test to check if the models.PublicKey has the "Data" field.
	return NewErrInvalid(ErrPublicKeyDataInvalid, map[string]interface{}{"Data": value}, next)
}

// NewErrPublicKeyFilter returns an error when the public key has more than one filter.
func NewErrPublicKeyFilter(next error) error {
	return NewErrInvalid(ErrPublicKeyFilter, nil, next)
}

// NewErrDeviceNotFound returns an error when the device is not found.
func NewErrDeviceNotFound(id models.UID, next error) error {
	return NewErrNotFound(ErrDeviceNotFound, string(id), next)
}

// NewErrSessionNotFound returns an error when the session is not found.
func NewErrSessionNotFound(id models.UID, next error) error {
	return NewErrNotFound(ErrSessionNotFound, string(id), next)
}

// NewErrNamespaceList return an error to be used when cannot list namespaces.
func NewErrNamespaceList(next error) error {
	return NewErrInvalid(ErrNamespaceList, nil, next)
}

// NewErrNamespaceInvalid returns an error to be used when the namespace is invalid.
func NewErrNamespaceInvalid(next error) error {
	return NewErrInvalid(ErrNamespaceInvalid, nil, next)
}

// NewErrNamespaceDuplicated returns an error to be used when the namespace is duplicated.
func NewErrNamespaceDuplicated(next error) error {
	return NewErrDuplicated(ErrNamespaceDuplicated, nil, next)
}

// NewErrNamespaceCreateStore returns an error to be used when the store function that create a namespace fails.
func NewErrNamespaceCreateStore(next error) error {
	return NewErrStore(ErrNamespaceCreateStore, nil, next)
}

// NewErrNamespaceMemberInvalid returns an error to be used when the namespace member is invalid.
func NewErrNamespaceMemberInvalid(next error) error {
	return NewErrInvalid(ErrNamespaceMemberInvalid, nil, next)
}

// NewErrNamespaceMemberNotFound returns an error to be used when the namespace member is not found.
func NewErrNamespaceMemberNotFound(id string, next error) error {
	return NewErrNotFound(ErrNamespaceMemberNotFound, id, next)
}

// NewErrNamespaceMemberFillData returns an error to be used when the conversion of models.Member, with only the ID and
// role set, to a complete structure, fails.
func NewErrNamespaceMemberFillData(next error) error {
	return NewErrInvalid(ErrNamespaceMemberFillData, nil, next)
}

// NewErrNamespaceMemberDuplicated returns an error to be used when the namespace member already exist in the namespace.
func NewErrNamespaceMemberDuplicated(id string, next error) error {
	return NewErrDuplicated(ErrNamespaceMemberDuplicated, []string{id}, next)
}

// NewErrDeviceInvalid returns an error to be used when the device data is invalid.
func NewErrDeviceInvalid(data map[string]interface{}, next error) error {
	return NewErrInvalid(ErrDeviceInvalid, data, next)
}

// NewErrDeviceDuplicated returns an error to be used when the device already exist in the namespace.
func NewErrDeviceDuplicated(name string, next error) error {
	return NewErrDuplicated(ErrDeviceDuplicated, []string{name}, next)
}

// NewErrDeviceLookupNotFound returns an error to be used when the device lookup is not found.
func NewErrDeviceLookupNotFound(namespace, name string, next error) error {
	return NewErrNotFound(ErrDeviceLookupNotFound, fmt.Sprintf("device %s on namespace %s", name, namespace), next)
}

// NewErrDeviceLimit returns an error to be used when the device limit is reached.
func NewErrDeviceLimit(limit int, next error) error {
	return NewErrLimit(ErrDeviceLimit, limit, next)
}

// NewErrDeviceStatusInvalid returns an error to be used when the device's status is invalid.
func NewErrDeviceStatusInvalid(status string, next error) error {
	return NewErrInvalid(ErrDeviceStatusInvalid, map[string]interface{}{"status": status}, next)
}

// NewErrDeviceStatusAccepted returns an error to be used when the device's status is accepted.
func NewErrDeviceStatusAccepted(next error) error {
	// This error is so tied to the device status, that it is not possible to use the NewErrInvalid function without this
	// literal assignment.
	return NewErrInvalid(ErrDeviceStatusAccepted, map[string]interface{}{"status": "accepted"}, next)
}

// NewErrTokenSigned returns an error to be used when the token signed fails.
func NewErrTokenSigned(err error) error {
	return NewErrInvalid(ErrTokenSigned, nil, err)
}

// NewErrUserNotConfirmed returns an error to be used when the user is not confirmed.
func NewErrUserNotConfirmed(err error) error {
	return NewErrForbidden(ErrUserNotConfirmed, err)
}

// NewErrAuthInvalid returns a error to be used when the auth data is invalid.
func NewErrAuthInvalid(data map[string]interface{}, err error) error {
	return NewErrInvalid(ErrAuthInvalid, data, err)
}

// NewErrUserUpdate returns a error to be used when the user update fails.
func NewErrUserUpdate(user *models.User, err error) error {
	return NewErrStore(ErrUserUpdate, user, err)
}

// NewErrDeviceCreate returns a error to be used when the device create fails.
func NewErrDeviceCreate(device models.Device, err error) error {
	return NewErrStore(ErrDeviceCreate, device, err)
}

// NewErrDeviceSetOnline returns a error to be used when the device set online fails.
func NewErrDeviceSetOnline(id models.UID, err error) error {
	return NewErrStore(ErrDeviceSetOnline, id, err)
}

// NewErrAuthUnathorized returns a error to be used when the auth is unauthorized.
func NewErrAuthUnathorized(err error) error {
	return NewErrUnathorized(ErrAuthUnathorized, err)
}

// NewErrBadRequest returns a error to be used when the auth is unauthorized.
func NewErrBadRequest(err error) error {
	return NewErrRequest(ErrBadRequest, err)
}

// NewErrNamespaceLimitReached a error to be used when the user namespace limit number is reached.
func NewErrNamespaceLimitReached(limit int, err error) error {
	return NewErrLimit(ErrNamespaceLimitReached, limit, err)
}

func NewErrDeviceRemovedCount(next error) error {
	return NewErrInvalid(ErrDeviceRemovedCount, nil, next)
}

func NewErrDeviceRemovedInsert(next error) error {
	return NewErrInvalid(ErrDeviceRemovedInsert, nil, next)
}

func NewErrDeviceRemovedFull(limit int, next error) error {
	return NewErrLimit(ErrDeviceRemovedFull, limit, next)
}

func NewErrDeviceRemovedDelete(next error) error {
	return NewErrInvalid(ErrDeviceRemovedDelete, nil, next)
}

func NewErrDeviceRemovedGet(next error) error {
	return NewErrInvalid(ErrDeviceRemovedGet, nil, next)
}

func NewErrBillingReportNamespaceDelete(next error) error {
	return NewErrInvalid(ErrBillingReportNamespaceDelete, nil, next)
}

func NewErrBillingReportDevice(next error) error {
	return NewErrInvalid(ErrBillingReportDevice, nil, next)
}

func NewErrBillingEvaluate(next error) error {
	return NewErrInvalid(ErrBillingEvaluate, nil, next)
}

func NewErrDeviceMaxDevicesReached(count int) error {
	return NewErrLimit(ErrMaxDeviceCountReached, count, nil)
}

func NewErrAuthForbidden() error {
	return NewErrForbidden(ErrAuthForbidden, nil)
}

func NewErrNamespaceIPInvalid() error {
	return NewErrInvalid(ErrNamespaceIPInvalid, nil, nil)
}

func NewErrNamespaceIPNotPrivate() error {
	return NewErrInvalid(ErrNamespaceIPNotPrivate, nil, nil)
}
