package services

import "github.com/shellhub-io/shellhub/pkg/envs"

// getMaxDevices get the limit of devices that a namespace can have if environment
// is cloud.
func getMaxDevices() int {
	if envs.IsCloud() {
		return MaxNumberDevicesLimited
	}

	return MaxNumberDevicesUnlimited
}
