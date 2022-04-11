package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
)

type SSHKeysTagsService interface {
	AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
	RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
	UpdatePublicKeyTags(ctx context.Context, tenant, fingerprint string, tags []string) error
}

// AddPublicKeyTag trys to add a tag to the models.PublicKey, when its filter is from Tags type.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the addition action.
func (s *service) AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	// Checks if the namespace exists.
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	// Checks if the public key exists.
	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return ErrPublicKeyNotFound
	}

	if key.Filter.Hostname != "" {
		return ErrPublicKeyInvalid
	}

	if len(key.Filter.Tags) == DeviceMaxTags {
		return ErrMaxTagReached
	}

	tags, _, err := s.store.TagsGet(ctx, tenant)
	if err != nil {
		return err
	}

	if !contains(tags, tag) {
		return ErrTagNameNotFound
	}

	// Trys to add a public key.
	err = s.store.PublicKeyAddTag(ctx, tenant, fingerprint, tag)
	if err != nil {
		switch err {
		case store.ErrNoDocuments:
			return ErrDuplicateTagName
		default:
			return err
		}
	}

	return nil
}

// RemovePublicKeyTag trys to remove a tag from the models.PublicKey, when its filter is from Tags type.
func (s *service) RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	// Checks if the namespace exists.
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	// Checks if the public key exists.
	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return ErrPublicKeyNotFound
	}

	if key.Filter.Hostname != "" {
		return ErrPublicKeyInvalid
	}

	// Checks if the tag already exists in the device.
	if !contains(key.Filter.Tags, tag) {
		return ErrTagNameNotFound
	}

	// Trys to remove a public key.
	err = s.store.PublicKeyRemoveTag(ctx, tenant, fingerprint, tag)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePublicKeyTags trys to update the tags of the models.PublicKey, when its filter is from Tags type.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the update action.
func (s *service) UpdatePublicKeyTags(ctx context.Context, tenant, fingerprint string, tags []string) error {
	set := func(list []string) []string {
		state := make(map[string]bool)
		helper := make([]string, 0)
		for _, item := range list {
			if _, ok := state[item]; !ok {
				state[item] = true
				helper = append(helper, item)
			}
		}

		return helper
	}

	// Checks if the namespace exists.
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	// Checks if the public key exists.
	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return ErrPublicKeyNotFound
	}

	if key.Filter.Hostname != "" {
		return ErrPublicKeyInvalid
	}

	if len(tags) > DeviceMaxTags {
		return ErrMaxTagReached
	}

	tags = set(tags)

	allTags, _, err := s.store.TagsGet(ctx, tenant)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if !contains(allTags, tag) {
			return ErrTagNameNotFound
		}
	}

	// Trys to add a public key.
	err = s.store.PublicKeyUpdateTags(ctx, tenant, fingerprint, tags)
	if err != nil {
		switch err {
		case store.ErrNoDocuments:
			return ErrDuplicateTagName
		default:
			return err
		}
	}

	return nil
}
