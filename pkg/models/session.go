package models

import (
	"time"
)

type SessionPosition struct {
	Longitude float64 `json:"longitude" bson:"longitude"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
}

type Session struct {
	UID           string          `json:"uid"`
	DeviceUID     UID             `json:"device_uid,omitempty" bson:"device_uid"`
	Device        *Device         `json:"device" bson:"-"`
	TenantID      string          `json:"tenant_id" bson:"tenant_id"`
	Username      string          `json:"username"`
	IPAddress     string          `json:"ip_address" bson:"ip_address"`
	StartedAt     time.Time       `json:"started_at" bson:"started_at"`
	LastSeen      time.Time       `json:"last_seen" bson:"last_seen"`
	Active        bool            `json:"active" bson:"active"`
	Closed        bool            `json:"-" bson:"closed"`
	Authenticated bool            `json:"authenticated" bson:"authenticated"`
	Recorded      bool            `json:"recorded" bson:"recorded"`
	Type          string          `json:"type" bson:"type"`
	Term          string          `json:"term" bson:"term"`
	Position      SessionPosition `json:"position" bson:"position"`
	Events        SessionEvents   `json:"events" bson:"events"`
}

type ActiveSession struct {
	UID      UID       `json:"uid"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	TenantID string    `json:"tenant_id" bson:"tenant_id"`
}

// NOTE: This struct has been moved to the cloud repo as it is only used in a cloud context;
// however, it is also utilized by migrations. For this reason, we must maintain the struct
// here ensure everything continues to function as expected.
// TODO: Remove this struct when it is no longer needed for migrations.
type RecordedSession struct {
	UID      UID       `json:"uid"`
	Message  string    `json:"message" bson:"message"`
	TenantID string    `json:"tenant_id" bson:"tenant_id,omitempty"`
	Time     time.Time `json:"time" bson:"time,omitempty"`
	Width    int       `json:"width" bson:"width,omitempty"`
	Height   int       `json:"height" bson:"height,omitempty"`
}

type Status struct {
	Authenticated bool `json:"authenticated"`
}

type SessionRecorded struct {
	UID       string `json:"uid"`
	Seat      int    `json:"seat"`
	Namespace string `json:"namespace" bson:"namespace"`
	Message   string `json:"message" bson:"message"`
	Width     int    `json:"width" bson:"width,omitempty"`
	Height    int    `json:"height" bson:"height,omitempty"`
}

type SessionUpdate struct {
	Authenticated *bool   `json:"authenticated"`
	Type          *string `json:"type"`
}

// SessionEvent represents a session event.
type SessionEvent struct {
	// Type of the session. Normally, it is the SSH request name.
	Type string `json:"type" bson:"type"`
	// Timestamp contains the time when the event was logged.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	// Data is a generic structure containing data of the event, normally the unmarshaling data of the request.
	Data any `json:"data" bson:"data"`
	// Seat is the seat where the event occurred.
	Seat int `json:"seat" bson:"seat"`
}

// SessionEvents stores the events registered in a session.
type SessionEvents struct {
	// Types field is a set of sessions type to simplify the indexing on the database.
	Types []string `json:"types" bson:"types,omitempty"`
	// Items contains a list of events happened in a session.
	Items []SessionEvent `json:"items" bson:"items,omitempty"`
	// Seats contains a list of seats of events.
	Seats []int `json:"seats" bson:"seats,omitempty"`
}

// SessionSeat stores a session's seat.
type SessionSeat struct {
	// ID is the identifier of session's seat.
	ID int `json:"id" bson:"id,omitempty"`
}
