package internal

import (
	"context"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Mongoclient is threadsafe inherently, no need for locking manually
type DBClient struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

type CrawledData struct {
	URL      string   `bson:"url"`
	Keywords []string `bson:"keywords"`
}

func (c *DBClient) InsertWebpage(job string, keywords []string) {
	_, err := c.Collection.InsertOne(context.TODO(), CrawledData{
		URL:      job,
		Keywords: keywords,
	})
	if err != nil {
		log.Printf("Failed to insert record, %v", err)
	}
}

func (c *DBClient) Disconnect() {
	if err := c.Client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

// Creating a basic inverted index (no ranking)
func (c *DBClient) CreateInvertedIndex() error {
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "keywords", Value: 1}},
		Options: options.Index().SetName("keyword_inverted_index"),
	}

	_, err := c.Collection.Indexes().CreateOne(context.TODO(), model)
	return err
}

// Search queries the inverted index for a specific keyword
func (c *DBClient) Search(keyword string) error {
	keyword = strings.ToLower(keyword)
	// This filter triggers the Multikey Index lookup
	filter := bson.D{{Key: "keywords", Value: keyword}}

	cursor, err := c.Collection.Find(context.TODO(), filter)
	if err != nil {
		return err
	}

	// Get all keyword matches (UNRANKED!!!)
	var results []CrawledData
	if err = cursor.All(context.TODO(), &results); err != nil {
		return err
	}

	// Print the results
	for _, res := range results {
		log.Printf("%s", res.URL)
	}
	return nil
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
	}
}
