package secret

import (
	"context"
	"sync"
	"time"

	"secretserver/external/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionName = "secrets"
)

var (
	coll     *mongo.Collection
	collOnce sync.Once
	collErr  error
)

// collection returns the secrets collection, connecting on first use.
// It no longer runs at package import time, so importing this package for a
// unit test does not require a live database.
func collection() (*mongo.Collection, error) {
	collOnce.Do(func() {
		database, err := mongodb.DB()
		if err != nil {
			collErr = err
			return
		}
		coll = database.Collection(CollectionName)
		ensureIndexes(coll)
	})
	return coll, collErr
}

// ensureIndexes creates a TTL index on expiresAt and a unique index on hash.
// The TTL index purges expired secrets server side. The unique index turns the
// hash identifier assumption into a guarantee. Failures are non-fatal.
func ensureIndexes(c *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
}
