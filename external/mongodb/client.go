package mongodb

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DBName = "secretserver"
)

var (
	client  *mongo.Client
	db      *mongo.Database
	once    sync.Once
	initErr error
)

// connect dials MongoDB once and caches the client. It runs on first use
// instead of at import time. That way importing this package does not crash
// a process, such as a unit test, when no database is reachable.
func connect() error {
	once.Do(func() {
		uri := os.Getenv("MONGO_URI")
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			initErr = err
			return
		}

		// Check the connection.
		if err := c.Ping(ctx, nil); err != nil {
			initErr = err
			return
		}

		client = c
		db = c.Database(DBName)

		fmt.Println("Connected to MongoDB!")
	})
	return initErr
}

// Client returns the shared MongoDB client, connecting on first call.
func Client() (*mongo.Client, error) {
	if err := connect(); err != nil {
		return nil, err
	}
	return client, nil
}

// DB returns the shared database handle, connecting on first call.
func DB() (*mongo.Database, error) {
	if err := connect(); err != nil {
		return nil, err
	}
	return db, nil
}
