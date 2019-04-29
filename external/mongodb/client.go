package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DBName = "secretserver"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

func init() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	c, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = c.Ping(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	client = c
	db = c.Database(DBName)

	fmt.Println("Connected to MongoDB!")
}

func Client() *mongo.Client {
	return client
}

func DB() *mongo.Database {
	return db
}
