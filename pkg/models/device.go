package models

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type DeviceStatus string

const (
	DeviceStatusAccepted DeviceStatus = "accepted"
	DeviceStatusPending  DeviceStatus = "pending"
	DeviceStatusRejected DeviceStatus = "rejected"
	DeviceStatusRemoved  DeviceStatus = "removed"
	DeviceStatusUnused   DeviceStatus = "unused"
	DeviceStatusEmpty    DeviceStatus = ""
)

type Device struct {
	// UID is the unique identifier for a device.
	UID string `json:"uid"`

	// ConnectedAt represents the timestamp (in UTC) of the device's most recent connection initiation with the server.
	// It does not determine the online status, of the device, use [Device.LastSeen] for this.
	ConnectedAt time.Time `json:"connected_at" bson:"connected_at"`

	// DisconnectedAt represents the timestamp (in UTC) of the device's most recent connection termination with the server.
	// It does not determine the online status of the device; use [Device.LastSeen] for this purpose.
	//
	// DisconnectedAt serves as a placeholder for [Device.LastSeen] (as it is not saved in the database). The last seen
	// should be set to this when it is not present in the cache.
	DisconnectedAt time.Time `json:"-" bson:"disconnected_at"`

	// LastSeen represents the timestamp (in UTC) of the device's latest ping to the server. It is utilized to determine
	// whether the device is currently online or offline.
	//
	// LastSeen data is not persisted in the database. Instead, it is exclusively stored in the cache and is merged
	// with the model during retrieval operations at the code level. When a device is offline, the value is not present in
	// the cache. In such cases, the value must be set to [Device.DisconnectedAt].
	LastSeen time.Time `json:"last_seen" bson:"-"`

	Name             string          `json:"name" bson:"name,omitempty" validate:"required,device_name"`
	Identity         *DeviceIdentity `json:"identity"`
	Info             *DeviceInfo     `json:"info"`
	PublicKey        string          `json:"public_key" bson:"public_key"`
	TenantID         string          `json:"tenant_id" bson:"tenant_id"`
	Online           bool            `json:"online" bson:",omitempty"`
	Namespace        string          `json:"namespace" bson:",omitempty"`
	Status           DeviceStatus    `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused"`
	StatusUpdatedAt  time.Time       `json:"status_updated_at" bson:"status_updated_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at" bson:"created_at,omitempty"`
	RemoteAddr       string          `json:"remote_addr" bson:"remote_addr"`
	Position         *DevicePosition `json:"position" bson:"position"`
	Tags             []string        `json:"tags" bson:"tags,omitempty"`
	PublicURL        bool            `json:"public_url" bson:"public_url,omitempty"`
	PublicURLAddress string          `json:"public_url_address" bson:"public_url_address,omitempty"`
	Acceptable       bool            `json:"acceptable" bson:"acceptable,omitempty"`
}

type DeviceChanges struct {
	ConnectedAt    time.Time `bson:"connected_at,omitempty"`
	DisconnectedAt time.Time `bson:"disconnected_at,omitempty"`
}

type DeviceAuthClaims struct {
	UID    string `json:"uid"`
	Tenant string `json:"tenant"`
	Status string `json:"status"`

	AuthClaims           `mapstruct:",squash"`
	jwt.RegisteredClaims `mapstruct:",squash"`
}

func (d *DeviceAuthClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	d.RegisteredClaims = claims
}

type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Hostname  string          `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" bson:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`
}

type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeviceIdentity struct {
	MAC string `json:"mac"`
}

type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

type DevicePosition struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type DeviceRemoved struct {
	Device    *Device   `json:"device" bson:"device"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type DeviceTag struct {
	Tag string `validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

func NewDeviceTag(tag string) DeviceTag {
	return DeviceTag{
		Tag: tag,
	}
}

// TODO:
type ConnectedDevice struct {
	UID      string `json:"uid"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	Status   string `json:"status" bson:"status"`
}
