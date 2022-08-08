package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) NamespaceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error) {
	query := []bson.M{}
	queryMatch, err := queries.BuildFilterQuery(filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	if export {
		query = []bson.M{
			{
				"$lookup": bson.M{
					"from":         "devices",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "devices",
				},
			},
			{
				"$addFields": bson.M{
					"devices": bson.M{"$size": "$devices"},
				},
			},
			{
				"$lookup": bson.M{
					"from":         "sessions",
					"localField":   "devices.uid",
					"foreignField": "device_uid",
					"as":           "sessions",
				},
			},
			{
				"$addFields": bson.M{
					"sessions": bson.M{"$size": "$sessions"},
				},
			},
		}
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if id := gateway.IDFromContext(ctx); id != nil {
		user, _, err := s.UserGetByID(ctx, id.ID, false)
		if err != nil {
			return nil, 0, err
		}

		query = append(query, bson.M{
			"$match": bson.M{
				"members": bson.M{
					"$elemMatch": bson.M{
						"id": user.ID,
					},
				},
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("namespaces"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page != 0 && pagination.PerPage != 0 {
		query = append(query, queries.BuildPaginationQuery(pagination)...)
	}

	namespaces := make([]models.Namespace, 0)
	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		err = cursor.Decode(&namespace)
		if err != nil {
			return namespaces, count, err
		}

		countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": namespace.TenantID, "status": "accepted"})
		if err != nil {
			return namespaces, 0, err
		}

		namespace.DevicesCount = int(countDevice)

		namespaces = append(namespaces, *namespace)
	}

	return namespaces, count, err
}

func (s *Store) NamespaceGet(ctx context.Context, tenantID string) (*models.Namespace, error) {
	var ns *models.Namespace

	if err := s.cache.Get(ctx, strings.Join([]string{"namespace", tenantID}, "/"), &ns); err != nil {
		logrus.Error(err)
	}

	if ns != nil {
		goto count
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		return ns, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"namespace", tenantID}, "/"), ns, time.Minute); err != nil {
		logrus.Error(err)
	}

count:
	countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": tenantID, "status": "accepted"})
	if err != nil {
		return nil, FromMongoError(err)
	}

	ns.DevicesCount = int(countDevice)

	return ns, nil
}

func (s *Store) NamespaceGetByName(ctx context.Context, name string) (*models.Namespace, error) {
	var ns *models.Namespace

	if err := s.cache.Get(ctx, strings.Join([]string{"namespace", name}, "/"), &ns); err != nil {
		logrus.Error(err)
	}

	if ns != nil {
		return ns, nil
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": name}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	_, err := s.db.Collection("namespaces").InsertOne(ctx, namespace)

	return namespace, err
}

func (s *Store) NamespaceDelete(ctx context.Context, tenantID string) error {
	if _, err := s.db.Collection("namespaces").DeleteOne(ctx, bson.M{"tenant_id": tenantID}); err != nil {
		return FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	collections := []string{"devices", "sessions", "connected_devices", "firewall_rules", "public_keys", "recorded_sessions"}
	for _, collection := range collections {
		if _, err := s.db.Collection(collection).DeleteMany(ctx, bson.M{"tenant_id": tenantID}); err != nil {
			return FromMongoError(err)
		}
	}

	return nil
}

func (s *Store) NamespaceRename(ctx context.Context, tenantID string, name string) (*models.Namespace, error) {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"name": namespace.Name, "max_devices": namespace.MaxDevices, "settings.session_record": namespace.Settings.SessionRecord}}); err != nil {
		return FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceAddMember(ctx context.Context, tenantID string, memberID string, memberRole string) (*models.Namespace, error) {
	result := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": memberID}}})
	if result.Err() == nil {
		return nil, ErrNamespaceDuplicatedMember
	}

	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$addToSet": bson.M{"members": bson.M{"id": memberID, "role": memberRole}}})
	if err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) (*models.Namespace, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"members": bson.M{"id": memberID}}})
	if err != nil {
		return nil, FromMongoError(err)
	}
	if result.ModifiedCount == 0 {
		return nil, ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceEditMember(ctx context.Context, tenantID string, memberID string, memberNewRole string) error {
	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID, "members.id": memberID}, bson.M{"$set": bson.M{"members.$.role": memberNewRole}})
	if err != nil {
		return FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceGetFirst(ctx context.Context, id string) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"settings.session_record": sessionRecord}}); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	var settings struct {
		Settings *models.NamespaceSettings `json:"settings" bson:"settings"`
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&settings); err != nil {
		return false, FromMongoError(err)
	}

	return settings.Settings.SessionRecord, nil
}
