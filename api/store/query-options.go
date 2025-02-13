package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type (
	NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error
	DeviceQueryOption    func(ctx context.Context, device *models.Device) error
	PublicKeyQueryOption func(ctx context.Context, publicKKey *models.PublicKey) error
)

type QueryOptions interface {
	// CountAcceptedDevices counts the devices with a status 'accepted'
	// in the namespace.
	CountAcceptedDevices() NamespaceQueryOption

	// EnrichMembersData join the user's data into members array.
	EnrichMembersData() NamespaceQueryOption

	// DeviceWithTagDetails join the tag's details into tags array.
	DeviceWithTagDetails() DeviceQueryOption

	// PublicKeyWithTagDetails join the tag's details into tags array.
	PublicKeyWithTagDetails() PublicKeyQueryOption
}
