package requests

import "github.com/shellhub-io/shellhub/pkg/api/query"

type CreateTag struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	Name     string `json:"name" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

type PushTag struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	Name     string `param:"name" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
	// TargetID is the identifier of the target to push the tag on.
	// For the reason cannot of it can be a list of things (UID for device, ID for firewall, etc...), it
	// cannot be parsed and must be set manually
	TargetID string `validate:"required"`
}

type PullTag struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	Name     string `param:"name" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
	// TargetID is the identifier of the target to pull the tag of.
	// For the reason cannot of it can be a list of things (UID for device, ID for firewall, etc...), it
	// cannot be parsed and must be set manually
	TargetID string `validate:"required"`
}

type ListTags struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	query.Paginator
	query.Filters
	query.Sorter
}

type UpdateTag struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	Name     string `param:"name" validate:"required"`
	// Similar to [UpdateTag.Name], but is used to update the tag's name instead of retrieve the tag.
	NewName string `json:"name" validate:"omitempty,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

type DeleteTag struct {
	TenantID string `param:"tenant" validate:"required,uuid"`
	Name     string `param:"name" validate:"required"`
}

// TagParam is a structure to represent and validate a tag as path param.
type TagParam struct {
	Tag string `param:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// TagBody is a structure to represent and validate a tag as json request body.
type TagBody struct {
	Tag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// TagDelete is the structure to represent the request data for delete tag endpoint.
type TagDelete struct {
	TagParam
}

// TagRename is the structure to represent the request data for rename tag endpoint.
type TagRename struct {
	TagParam
	NewTag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}
