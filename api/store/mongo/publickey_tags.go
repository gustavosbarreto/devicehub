package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) PublicKeyPushTag(ctx context.Context, tenant, fingerprint, tag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer session.EndSession(ctx)

	_, erro := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		if _, err := s.TagGet(sessCtx, tag, tenant); err != nil {
			if errors.Is(err, store.ErrNoDocuments) {
				err := s.TagsPushTag(sessCtx, tag, tenant)
				if err != nil {
					return nil, FromMongoError(err)
				}
			} else if err != nil {
				return nil, err
			}
		}

		result, err := s.db.Collection("public_keys").
			UpdateOne(sessCtx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint},
				bson.M{"$addToSet": bson.M{"filter.tags": tag}})
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		return nil, nil
	})

	if erro != nil {
		return erro
	}

	return nil
}

func (s *Store) PublicKeyPullTag(ctx context.Context, tenant, fingerprint, tag string) error {
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) PublicKeySetTags(ctx context.Context, tenant, fingerprint string, tags []string) (int64, int64, error) {
	res, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$set": bson.M{"filter.tags": tags}})

	return res.MatchedCount, res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTag}, bson.M{"$set": bson.M{"filter.tags.$": newTag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("public_keys").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), FromMongoError(err)
}
