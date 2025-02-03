package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) Options() store.QueryOptions {
	return s.options
}

func (*queryOptions) CountAcceptedDevices() store.NamespaceQueryOption {
	return func(ctx context.Context, ns *models.Namespace) error {
		db, ok := ctx.Value("db").(*mongo.Database)
		if !ok {
			return errors.New("db not found in context")
		}

		countDevice, err := db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": ns.TenantID, "status": "accepted"})
		if err != nil {
			return FromMongoError(err)
		}

		ns.DevicesCount = int(countDevice)

		return nil
	}
}

func (*queryOptions) EnrichMembersData() store.NamespaceQueryOption {
	return func(ctx context.Context, ns *models.Namespace) error {
		db, ok := ctx.Value("db").(*mongo.Database)
		if !ok {
			return errors.New("db not found in context")
		}

		for i, member := range ns.Members {
			user := new(models.User)
			objID, _ := primitive.ObjectIDFromHex(member.ID)

			if err := db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
				log.WithError(err).
					WithField("id", member.ID).
					Error("member not found")

				continue
			}

			ns.Members[i].Email = user.Email
		}

		return nil
	}
}

func (*queryOptions) DeviceWithTagDetails() store.DeviceQueryOption {
	return func(ctx context.Context, device *models.Device) error {
		s, ok := ctx.Value("store").(store.Store)
		if !ok {
			return errors.New("store not found in context")
		}

		device.Tags = []models.Tag{}
		for _, tagID := range device.TagsID {
			tag, err := s.TagGetByID(ctx, tagID)
			if err != nil {
				log.WithError(err).WithField("id", tagID).Error("cannot retrieve tag")

				continue
			}

			device.Tags = append(device.Tags, *tag)
		}

		return nil
	}
}

func (*queryOptions) PublicKeyWithTagDetails() store.PublicKeyQueryOption {
	return func(ctx context.Context, publicKey *models.PublicKey) error {
		s, ok := ctx.Value("store").(store.Store)
		if !ok {
			return errors.New("store not found in context")
		}

		publicKey.Filter.Tags = []models.Tag{}
		for _, tagID := range publicKey.Filter.TagsID {
			tag, err := s.TagGetByID(ctx, tagID)
			if err != nil {
				log.WithError(err).WithField("id", tagID).Error("cannot retrieve tag")

				continue
			}

			publicKey.Filter.Tags = append(publicKey.Filter.Tags, *tag)
		}

		return nil
	}
}
