package internal

import (
	"log"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBClient struct {
	Client *mongo.Client
	Mu     *sync.Mutex
}

func ConnectMongo() DBClient {
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to estabilish connection")
	}

	return DBClient{
		Client: client,
		Mu:     &sync.Mutex{},
	}
}
