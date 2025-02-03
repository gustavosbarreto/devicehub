package models

import "time"

// Taggable is an embeddable struct that adds tagging capability to other models.
//
// Example usage:
//
//	type Device struct {
//	    Taggable    // Embed the Taggable struct
//	    Name string // Other device fields
//	}
type Taggable struct {
	// TagsID contains the IDs of associated tags. It is used only for database storage
	// and relationship management. The field is not exposed in JSON responses to keep
	// the API focused on meaningful tag data rather than internal identifiers.
	TagsID []string `json:"-" bson:"tags,omitempty"`

	// Tags contains the complete Tag objects associated with this resource. This field
	// is populated from TagsID when retrieving data from the database, but is not
	// stored directly. It is used only for JSON serialization to provide clients
	// with full tag information.
	Tags []Tag `json:"tags,omitempty" bson:"-"`
}

type Tag struct {
	ID        string    `json:"-" bson:"_id"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`
	Name      string    `json:"name" bson:"name"`
	TenantID  string    `json:"-" bson:"tenant_id"`
}

type TagChanges struct {
	Name string `bson:"name,omitempty"`
}

type TagConflicts struct {
	Name string
}

type TagTarget int

const (
	TagTargetDevice TagTarget = iota + 1
	TagTargetPublicKey
	TagTargetFirewallRule
)

func TagTargets() []TagTarget {
	return []TagTarget{TagTargetDevice, TagTargetPublicKey, TagTargetFirewallRule}
}
