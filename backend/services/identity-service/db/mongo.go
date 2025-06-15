package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDatabase *mongo.Database

// InitDB initializes the database connection
func InitDB() error {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("MONGO_URI environment variable not set, using default: mongodb://localhost:27017/gestor_e_docs")
		mongoURI = "mongodb://localhost:27017/gestor_e_docs"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		log.Println("MONGO_DB_NAME environment variable not set, using default: gestor_e_docs")
		dbName = "gestor_e_docs"
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v\n", err)
		return err
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Error pinging MongoDB: %v\n", err)
		return err
	}

	MongoClient = client
	MongoDatabase = client.Database(dbName)

	log.Println("Successfully connected to MongoDB!")
	return nil
}

// GetCollection returns a collection from the database
func GetCollection(collectionName string) *mongo.Collection {
	if MongoDatabase == nil {
		// This should ideally not happen if InitDB is called at application startup
		// and its error is handled.
		log.Fatal("Database not initialized. Call InitDB first.")
		return nil
	}
	return MongoDatabase.Collection(collectionName)
}

// DisconnectDB closes the MongoDB connection
func DisconnectDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v\n", err)
		} else {
			log.Println("Successfully disconnected from MongoDB.")
		}
	}
}

// GetDatabase returns the MongoDB database instance
func GetDatabase() *mongo.Database {
	if MongoDatabase == nil {
		// Este erro não deveria ocorrer se InitDB for chamado na inicialização
		log.Fatal("Database not initialized. Call InitDB first.")
		return nil
	}
	return MongoDatabase
}
