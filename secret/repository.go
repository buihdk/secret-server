package secret

import (
	"secretserver/external/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CollectionName = "secrets"
)

var (
	db *mongo.Collection
)

func init() {
	db = mongodb.DB().Collection(CollectionName)
}
