package internal

import (
	"context"
	"log"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBClient struct {
	Client     *mongo.Client
	Collection *mongo.Collection
	Mu         *sync.Mutex
}

type CrawledData struct {
	URL      string
	Keywords []string
}

func (c *DBClient) InsertWebpage(job string, keywords []string) {
	c.Mu.Lock()
	_, err := c.Collection.InsertOne(context.TODO(), CrawledData{
		URL:      job,
		Keywords: keywords,
	})
	if err != nil {
		log.Fatalf("Failed to insert record, %v", err)
	}
	c.Mu.Unlock()
}

func (c *DBClient) Disconnect() {
	if err := c.Client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func ConnectMongo() DBClient {
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to estabilish connection")
	}

	return DBClient{
		Client:     client,
		Collection: client.Database("crawler").Collection("crawled-data"),
		Mu:         &sync.Mutex{},
	}
}
