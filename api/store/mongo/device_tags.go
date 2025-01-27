package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func (s *Store) DevicePushTag(ctx context.Context, uid models.UID, tag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer session.EndSession(ctx)

	_, erro := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		device := new(models.Device)

		err := s.db.Collection("devices", options.Collection().SetWriteConcern(writeconcern.Majority())).
			FindOne(sessCtx, bson.M{"uid": uid}).
			Decode(device)
		if err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.TagGet(sessCtx, tag, device.TenantID); err != nil {
			if errors.Is(err, store.ErrNoDocuments) {
				err := s.TagsPushTag(sessCtx, tag, device.TenantID)
				if err != nil {
					return nil, FromMongoError(err)
				}
			} else if err != nil {
				return nil, err
			}
		}

		t, err := s.db.Collection("devices", options.Collection().SetWriteConcern(writeconcern.Majority())).
			UpdateOne(sessCtx, bson.M{"uid": uid}, bson.M{"$push": bson.M{"tags": tag}})
		if err != nil {
			return nil, FromMongoError(err)
		}

		if t.ModifiedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		return nil, nil
	})

	return erro
}

func (s *Store) DevicePullTag(ctx context.Context, uid models.UID, tag string) error {
	t, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$pull": bson.M{"tags": tag}})
	if err != nil {
		return FromMongoError(err)
	}

	if t.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) DeviceSetTags(ctx context.Context, uid models.UID, tags []string) (int64, int64, error) {
	tag, err := s.db.Collection("devices").
		UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})

	if tag.MatchedCount < 1 {
		return tag.MatchedCount, tag.ModifiedCount, store.ErrNoDocuments
	}

	return tag.MatchedCount, tag.ModifiedCount, FromMongoError(err)
}

func (s *Store) DeviceBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("devices").
		UpdateMany(ctx, bson.M{"tenant_id": tenant, "tags": currentTag}, bson.M{"$set": bson.M{"tags.$": newTag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) TagsBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("tags", options.Collection().SetWriteConcern(writeconcern.Majority())).
		UpdateOne(ctx, bson.M{"tenant_id": tenant, "name": currentTag}, bson.M{"$set": bson.M{"name": newTag}})
	if err != nil {
		return 0, FromMongoError(err)
	}

	return res.ModifiedCount, err
}

func (s *Store) DeviceBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("devices").
		UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) DeviceGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("devices").Distinct(ctx, "tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), FromMongoError(err)
}
