package inputs

// NamespaceCreate defines the structure for inputs when creating a namespace.
type NamespaceCreate struct {
	Namespace string `validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	Owner     string `validate:"required,username"`
	TenantID  string `validate:"omitempty,uuid"`
}

// NamespaceDelete defines the structure for inputs when deleting a namespace.
type NamespaceDelete struct {
	Namespace string
}
