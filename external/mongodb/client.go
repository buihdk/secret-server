package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
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
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = c.Ping(ctx, nil)
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
