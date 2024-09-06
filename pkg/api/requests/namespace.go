package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

// TenantParam is a structure to represent and validate a namespace tenant as path param.
type TenantParam struct {
	Tenant string `param:"tenant" validate:"required,uuid"`
}

// RoleBody is a structure to represent and validate a namespace role as request body.
type RoleBody struct {
	Role string `json:"role" validate:"required,oneof=administrator operator observer"`
}

// MemberParam is a structure to represent and validate a member UID as path param.
type MemberParam struct {
	MemberUID string `param:"uid" validate:"required"`
}

// NamespaceCreate is the structure to represent the request data for create namespace endpoint.
type NamespaceList struct {
	query.Paginator
	query.Filters
}

// NamespaceCreate is the structure to represent the request data for create namespace endpoint.
type NamespaceCreate struct {
	UserID   string `header:"X-ID" validate:"required"`
	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	TenantID string `json:"tenant" validate:"omitempty,uuid"`
}

// NamespaceGet is the structure to represent the request data for get namespace endpoint.
type NamespaceGet struct {
	TenantParam
}

// NamespaceDelete is the structure to represent the request data for delete namespace endpoint.
type NamespaceDelete struct {
	TenantParam
}

// NamespaceEdit is the structure to represent the request data for edit namespace endpoint.
type NamespaceEdit struct {
	TenantParam
	Name     string `json:"name" validate:"omitempty,hostname_rfc1123,excludes=."`
	Settings struct {
		SessionRecord          *bool   `json:"session_record" validate:"omitempty"`
		ConnectionAnnouncement *string `json:"connection_announcement" validate:"omitempty,min=0,max=4096"`
	} `json:"settings"`
	VPN struct {
		// Enable defines if the Virtual Private Network between devices are enabled.
		Enable *bool `json:"enable"`
		// Address defines the network address.
		Address *[4]byte `json:"address"`
		// Mask defines the mask of the network.
		Mask *byte `json:"mask" validate:"omitempty,min=8,max=24"`
	} `json:"vpn"`
}

type NamespaceAddMember struct {
	UserID      string          `header:"X-ID" validate:"required"`
	TenantID    string          `param:"tenant" validate:"required,uuid"`
	MemberEmail string          `json:"email" validate:"required"`
	MemberRole  authorizer.Role `json:"role" validate:"required,member_role"`
}

type NamespaceUpdateMember struct {
	UserID     string          `header:"X-ID" validate:"required"`
	TenantID   string          `param:"tenant" validate:"required,uuid"`
	MemberID   string          `param:"uid" validate:"required"`
	MemberRole authorizer.Role `json:"role" validate:"omitempty,member_role"`
}

type NamespaceRemoveMember struct {
	UserID   string `header:"X-ID" validate:"required"`
	TenantID string `param:"tenant" validate:"required,uuid"`
	MemberID string `param:"uid" validate:"required"`
}

// SessionEditRecordStatus is the structure to represent the request data for edit session record status endpoint.
type SessionEditRecordStatus struct {
	TenantParam
	SessionRecord bool `json:"session_record"`
}
