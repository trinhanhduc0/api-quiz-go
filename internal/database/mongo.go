package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance        *mongo.Client
	clientOnce            sync.Once
	MongoConnectionString = "MONGODB_URI"
)

// ConnectMongoDB khởi tạo và trả về kết nối MongoDB
func ConnectMongoDB(uri string) *mongo.Client {
	clientOnce.Do(func() {
		clientOptions := options.Client().ApplyURI(uri)
		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			log.Fatalf("Failed to create MongoDB client: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Kiểm tra kết nối
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		clientInstance = client
		fmt.Println("Connected to MongoDB!")
	})

	return clientInstance
}

// GetMongoClient trả về client MongoDB đã kết nối
func GetMongoClient() *mongo.Client {
	if clientInstance == nil {
		log.Fatal("MongoDB client is not initialized")
	}
	return clientInstance
}
