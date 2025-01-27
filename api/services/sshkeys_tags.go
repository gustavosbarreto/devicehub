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
	if _, err := s.store.NamespaceGet(ctx, tenant); err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	// Checks if the public key exists.
	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return NewErrPublicKeyNotFound(fingerprint, err)
	}

	if key.Filter.Hostname != "" {
		return NewErrPublicKeyFilter(nil)
	}

	if len(key.Filter.Tags) == DeviceMaxTags {
		return NewErrTagLimit(DeviceMaxTags, nil)
	}

	tags, _, err := s.store.TagsGet(ctx, tenant)
	if err != nil {
		return NewErrTagEmpty(tenant, err)
	}

	if !containsTags(tags, tag) {
		return NewErrTagNotFound(tag, nil)
	}

	// Trys to add a public key.
	err = s.store.PublicKeyPushTag(ctx, tenant, fingerprint, tag)
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
	if _, err := s.store.NamespaceGet(ctx, tenant); err != nil {
		return NewErrNamespaceNotFound(tenant, nil)
	}

	// Checks if the public key exists.
	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return NewErrPublicKeyNotFound(fingerprint, err)
	}

	if key.Filter.Hostname != "" {
		return NewErrPublicKeyFilter(nil)
	}

	// Checks if the tag already exists in the device.
	if !contains(key.Filter.Tags, tag) {
		return NewErrTagNotFound(tag, nil)
	}

	// Trys to remove a public key.
	err = s.store.PublicKeyPullTag(ctx, tenant, fingerprint, tag)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePublicKeyTags trys to update the tags of the models.PublicKey, when its filter is from Tags type.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the update action.
func (s *service) UpdatePublicKeyTags(ctx context.Context, tenant, fingerprint string, tags []string) error {
	if len(tags) > DeviceMaxTags {
		return NewErrTagLimit(DeviceMaxTags, nil)
	}

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

	tags = set(tags)

	if _, err := s.store.NamespaceGet(ctx, tenant); err != nil {
		return NewErrNamespaceNotFound(tenant, nil)
	}

	key, err := s.store.PublicKeyGet(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return NewErrPublicKeyNotFound(fingerprint, err)
	}

	if key.Filter.Hostname != "" {
		return NewErrPublicKeyNotFound(fingerprint, nil)
	}

	allTags, _, err := s.store.TagsGet(ctx, tenant)
	if err != nil {
		return NewErrTagEmpty(tenant, err)
	}

	for _, tag := range tags {
		if !containsTags(allTags, tag) {
			return NewErrTagNotFound(tag, nil)
		}
	}

	if _, _, err := s.store.PublicKeySetTags(ctx, tenant, fingerprint, tags); err != nil {
		return err
	}

	return nil
}
