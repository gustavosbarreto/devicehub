package requests

import "time"

// SessionIDParam is a structure to represent and validate a session UID as path param.
type SessionIDParam struct {
	// UID is the session's UID.
	UID string `param:"uid" validate:"required"`
}

// SessionGet is the structure to represent the request data for get session endpoint.
type SessionGet struct {
	SessionIDParam
}

// SessionAuthenticatedSet is the structure to represent the request data for set authenticated session endpoint.
type SessionAuthenticatedSet struct {
	SessionIDParam
	Authenticated bool `json:"authenticated" validate:"required"`
}

// SessionCreate is the structure to represent the request data for create session endpoint.
type SessionCreate struct {
	UID       string `json:"uid" validate:"required"`
	DeviceUID string `json:"device_uid" validate:"required"`
	Username  string `json:"username" validate:"required"`
	IPAddress string `json:"ip_address" validate:"required"`
	Type      string `json:"type" validate:"required"`
	Term      string `json:"term" validate:""`
}

// SessionFinish is the structure to represent the request data for finish session endpoint.
type SessionFinish struct {
	SessionIDParam
}

// SessionFinish is the structure to represent the request data for keep alive session endpoint.
type SessionKeepAlive struct {
	SessionIDParam
}

type SessionUpdate struct {
	SessionIDParam
	Authenticated *bool   `json:"authenticated"`
	Type          *string `json:"type"`
}

type SessionEvent struct {
	SessionIDParam
	Type      string    `json:"type" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
	Data      any       `json:"data" validate:"required"`
	Seat      int       `json:"seat" validate:"min=0"`
}

type SessionSeat struct {
	SessionIDParam
	ID int `json:"id" bson:"id,omitempty"`
}
