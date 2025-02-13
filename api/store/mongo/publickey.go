package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) PublicKeyGet(ctx context.Context, fingerprint string, tenantID string, opts ...store.PublicKeyQueryOption) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)
	if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenantID}).Decode(&pubKey); err != nil {
		return nil, FromMongoError(err)
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "store", s), pubKey); err != nil { //nolint:revive
			return nil, err
		}
	}

	return pubKey, nil
}

func (s *Store) PublicKeyList(ctx context.Context, paginator query.Paginator, opts ...store.PublicKeyQueryOption) ([]models.PublicKey, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("public_keys"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, queries.FromPaginator(&paginator)...)

	list := make([]models.PublicKey, 0)
	cursor, err := s.db.Collection("public_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		key := new(models.PublicKey)
		err = cursor.Decode(&key)
		if err != nil {
			return list, count, err
		}

		ctx := context.WithValue(ctx, "store", s) //nolint:revive
		for _, opt := range opts {
			if err := opt(ctx, key); err != nil {
				return nil, 0, err
			}
		}

		list = append(list, *key)
	}

	return list, count, err
}

func (s *Store) PublicKeyCreate(ctx context.Context, key *models.PublicKey) error {
	_, err := s.db.Collection("public_keys").InsertOne(ctx, key)

	return FromMongoError(err)
}

func (s *Store) PublicKeyUpdate(ctx context.Context, fingerprint string, tenantID string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.M{"fingerprint": fingerprint, "tenant_id": tenantID}

	pubKey := new(models.PublicKey)
	if err := s.db.Collection("public_keys").FindOneAndUpdate(ctx, filter, bson.M{"$set": key}, opts).Decode(&pubKey); err != nil {
		return nil, FromMongoError(err)
	}

	return pubKey, nil
}

func (s *Store) PublicKeyDelete(ctx context.Context, fingerprint string, tenantID string) error {
	pubKey, err := s.db.Collection("public_keys").DeleteOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenantID})
	if err != nil {
		return FromMongoError(err)
	}

	if pubKey.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
