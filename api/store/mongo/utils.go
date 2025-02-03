package mongo

import (
	"context"
	stderrors "errors"
	"io"
	"reflect"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AggregateCount takes a pipeline and count the results.
func AggregateCount(ctx context.Context, coll *mongo.Collection, pipeline []bson.M) (int, error) {
	resp := struct {
		Count int `bson:"count"`
	}{}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, nil
	}

	if err = cursor.Decode(&resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for mongo's error.
const ErrLayer = "mongo"

// ErrMongo is the error for any unknown mongo error.
var ErrMongo = errors.New("mongo error", ErrLayer, 1)

func FromMongoError(err error) error {
	switch {
	case err == mongo.ErrNoDocuments, err == io.EOF:
		return store.ErrNoDocuments
	case err == primitive.ErrInvalidHex:
		return store.ErrInvalidHex
	case mongo.IsDuplicateKeyError(err):
		return store.ErrDuplicate
	default:
		if err == nil {
			return nil
		}

		return errors.Wrap(ErrMongo, err)
	}
}

// structToBson converts a struct to it's bson representation.
func structToBson[T any](v T) primitive.M {
	data, err := bson.Marshal(v)
	if err != nil {
		panic(err)
	}

	doc := make(primitive.M)
	if err := bson.Unmarshal(data, &doc); err != nil {
		panic(err)
	}

	return doc
}

// sanitizeBson recursively sanitizes a bson, setting zero-value fields to nil
func sanitizeBson(data primitive.M) {
	for k, v := range data {
		if reflect.TypeOf(v) == reflect.TypeOf(primitive.M{}) {
			sanitizeBson(v.(primitive.M))
		} else {
			if v != nil && reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface()) {
				data[k] = nil
			}
		}
	}
}

func collectionFromTagTarget(target models.TagTarget) (string, string, string, error) {
	switch target {
	case models.TagTargetDevice:
		return "devices", "uid", "tags", nil
	case models.TagTargetPublicKey:
		return "public_keys", "fingerprint", "filter.tags", nil
	case models.TagTargetFirewallRule:
		return "firewall_rules", "_id", "", nil
	default:
		return "", "", "", stderrors.New("invalid tag target")
	}
}
