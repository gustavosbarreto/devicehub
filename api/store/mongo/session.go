package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) SessionList(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error) {
	query := []bson.M{
		{
			"$match": bson.M{
				"uid": bson.M{
					"$ne": nil,
				},
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
	count, err := AggregateCount(ctx, s.db.Collection("sessions"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, bson.M{
		"$sort": bson.M{
			"started_at": -1,
		},
	})

	query = append(query, queries.FromPaginator(&paginator)...)
	query = append(query, []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
	}...)

	sessions := make([]models.Session, 0)
	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return sessions, count, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		session := new(models.Session)
		err = cursor.Decode(&session)
		if err != nil {
			return sessions, count, err
		}

		device, err := s.DeviceGet(ctx, session.DeviceUID)
		if err != nil {
			return sessions, count, err
		}

		session.Device = device
		sessions = append(sessions, *session)
	}

	return sessions, count, err
}

func (s *Store) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
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

	session := new(models.Session)

	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&session)
	if err != nil {
		return nil, FromMongoError(err)
	}

	device, err := s.DeviceGet(ctx, session.DeviceUID)
	if err != nil {
		return nil, FromMongoError(err)
	}

	session.Device = device

	return session, nil
}

func (s *Store) SessionUpdate(ctx context.Context, uid models.UID, model *models.Session) error {
	result, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": model})
	if err != nil {
		return FromMongoError(err)
	}

	if result.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error {
	session, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"recorded": recorded}})
	if err != nil {
		return FromMongoError(err)
	}

	if session.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error) {
	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := s.DeviceGet(ctx, session.DeviceUID)
	if err != nil {
		return nil, FromMongoError(err)
	}

	session.TenantID = device.TenantID

	if _, err := s.db.Collection("sessions").InsertOne(ctx, &session); err != nil {
		return nil, FromMongoError(err)
	}

	return &session, nil
}

func (s *Store) SessionSetLastSeen(ctx context.Context, uid models.UID) error {
	session := models.Session{}

	err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session)
	if err != nil {
		return FromMongoError(err)
	}

	if session.Closed {
		return nil
	}

	session.LastSeen = clock.Now()

	opts := options.Update().SetUpsert(true)
	_, err = s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return FromMongoError(err)
	}

	if _, err := s.db.Collection("active_sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"last_seen": clock.Now()}}); err != nil {
		return FromMongoError(err)
	}

	return nil
}

// SessionDeleteActives sets a session's "closed" status to true and deletes all related active_sessions.
func (s *Store) SessionDeleteActives(ctx context.Context, uid models.UID) error {
	mongoSession, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer mongoSession.EndSession(ctx)

	_, err = mongoSession.WithTransaction(ctx, func(_ mongo.SessionContext) (interface{}, error) {
		session := new(models.Session)

		query := bson.M{"uid": uid}
		update := bson.M{"$set": bson.M{"last_seen": clock.Now(), "closed": true}}

		if err := s.db.Collection("sessions").FindOneAndUpdate(ctx, query, update).Decode(&session); err != nil {
			return nil, FromMongoError(err)
		}

		_, err := s.db.Collection("active_sessions").DeleteMany(ctx, bson.M{"uid": session.UID})

		return nil, FromMongoError(err)
	})

	return err
}

func (s *Store) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	session, err := s.db.Collection("sessions").UpdateMany(ctx, bson.M{"device_uid": oldUID}, bson.M{"$set": bson.M{"device_uid": newUID}})
	if err != nil {
		return FromMongoError(err)
	}

	if session.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionActiveCreate(ctx context.Context, uid models.UID, session *models.Session) error {
	_, err := s.db.Collection("active_sessions").InsertOne(ctx, &models.ActiveSession{
		UID:      uid,
		LastSeen: session.StartedAt,
		TenantID: session.TenantID,
	})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

// SessionEvent saves a [models.SessionEvent] into the database.
//
// It pushes the event into events type array, and the event type into a separated set. The set is used to improve the
// performance of indexing when looking for sessions.
func (s *Store) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	if _, err := s.db.Collection("sessions").UpdateOne(ctx,
		bson.M{"uid": uid},
		bson.M{
			"$addToSet": bson.M{
				"events.types": event.Type,
				"events.seats": event.Seat,
			},
			"$push": bson.M{"events.items": event},
		},
	); err != nil {
		return FromMongoError(err)
	}

	return nil
}
