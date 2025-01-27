package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func (s *Store) FirewallRuleGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("firewall_rules").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), FromMongoError(err)
}

func (s *Store) FirewallRuleBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTag}, bson.M{"$set": bson.M{"filter.tags.$": newTag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) TagsRename(ctx context.Context, tenantID string, oldTag string, newTag string) (int64, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return int64(0), FromMongoError(err)
	}
	defer session.EndSession(ctx)

	count, err := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		devCount, err := s.DeviceBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		keyCount, err := s.PublicKeyBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		rulCount, err := s.FirewallRuleBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		tagsCount, err := s.TagsBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		return devCount + keyCount + rulCount + tagsCount, nil
	})
	if err != nil {
		return int64(0), FromMongoError(err)
	}

	return count.(int64), nil
}

func (s *Store) FirewallRuleBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) TagsDelete(ctx context.Context, tenantID string, tag string) (int64, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return int64(0), FromMongoError(err)
	}
	defer session.EndSession(ctx)

	count, err := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		devCount, err := s.DeviceBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		keyCount, err := s.PublicKeyBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		rulCount, err := s.FirewallRuleBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		tagCount, err := s.TagsBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		return devCount + keyCount + rulCount + tagCount, nil
	})
	if err != nil {
		return int64(0), FromMongoError(err)
	}

	return count.(int64), nil
}

func (s *Store) TagGet(ctx context.Context, tagName, tenant string) (*models.Tags, error) {
	tag := new(models.Tags)
	if err := s.db.Collection("tags").FindOne(ctx, bson.M{"name": tagName, "tenant_id": tenant}).Decode(tag); err != nil {
		return nil, FromMongoError(err)
	}

	return tag, nil
}

func (s *Store) TagsGet(ctx context.Context, tenant string) ([]models.Tags, int64, error) {
	tags, length, err := s.TagsGetTags(ctx, tenant)
	if err != nil {
		return nil, length, FromMongoError(err)
	}

	return removeDuplicate[models.Tags](tags), length, nil
}

func (s *Store) TagsPushTag(ctx context.Context, tagName, tenantID string) error {
	tag := &models.Tags{
		Name:   tagName,
		Tenant: tenantID,
		Color:  "",
	}

	_, err := s.db.Collection("tags", options.Collection().SetWriteConcern(writeconcern.Majority())).
		InsertOne(ctx, tag)
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) TagsBulkDeleteTag(ctx context.Context, tenant, tagName string) (int64, error) {
	res, err := s.db.Collection("tags", options.Collection().SetWriteConcern(writeconcern.Majority())).
		DeleteOne(ctx, bson.M{"tenant_id": tenant, "name": tagName})

	return res.DeletedCount, FromMongoError(err)
}

func (s *Store) TagsGetTags(ctx context.Context, tenant string) ([]models.Tags, int64, error) {
	cursor, err := s.db.Collection("tags").Find(ctx, bson.M{"tenant_id": tenant})
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	tags := make([]models.Tags, cursor.RemainingBatchLength())
	i := 0

	for cursor.Next(ctx) {
		tg := new(models.Tags)

		if err := cursor.Decode(tg); err != nil {
			return nil, int64(0), FromMongoError(err)
		}

		tags[i] = *tg //nolint:forcetypeassert

		i++
	}

	return tags, int64(len(tags)), FromMongoError(err)
}
