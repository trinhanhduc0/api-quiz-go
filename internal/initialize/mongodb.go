package initialize

import (
	"my_app/internal/database"
	"os"
)

func InitMongoDB() {
	// Get MongoDB connection string
	connectString := os.Getenv(database.MongoConnectionString)
	// Connect to MongoDB
	database.ConnectMongoDB(connectString)
}
