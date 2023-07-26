package envs

const (
	ENABLED = "true"
)

// Backend is an interface for any sort of underlying key/value store.
type Backend interface {
	Get(key string) string
}

// DefaultBackend define the backend to be used to get environment variables.
var DefaultBackend Backend

func init() {
	DefaultBackend = &envBackend{}
}

// IsEnterprise returns true if the current ShellHub server instance is enterprise.
func IsEnterprise() bool {
	return DefaultBackend.Get("SHELLHUB_ENTERPRISE") == ENABLED
}

// IsCloud returns true if the current ShellHub server instance is cloud.
func IsCloud() bool {
	return DefaultBackend.Get("SHELLHUB_CLOUD") == ENABLED
}

// HasBilling returns true if the current ShellHub server instance has billing feature enabled.
func HasBilling() bool {
	return DefaultBackend.Get("SHELLHUB_BILLING") == ENABLED
}

// IsCommunity return true if the current ShellHub server instance is community.
// It evaluates if the current ShellHub instance is neither enterprise or cloud .
func IsCommunity() bool {
	return (DefaultBackend.Get("SHELLHUB_CLOUD") != ENABLED && DefaultBackend.Get("SHELLHUB_ENTERPRISE") != ENABLED)
}
