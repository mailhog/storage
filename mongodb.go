package storage

import (
	"context"
	"time"

	"github.com/mailhog/data"

	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents MongoDB backed storage backend
type MongoDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

// CreateMongoDB creates a MongoDB backed storage backend
func CreateMongoDB(uri, db, coll string) *MongoDB {
	log.Printf("Connecting to MongoDB: %s\n", uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Error connecting to MongoDB: %s", err)
		return nil
	}

	mod := mongo.IndexModel{
		Keys:    bson.M{"created": 1}, // index in ascending order or -1 for descending order
		Options: options.Index().SetUnique(true),
	}

	_, err = mongoClient.Database(db).Collection(coll).Indexes().CreateOne(ctx, mod)
	if err != nil {
		log.Printf("Failed creating index: %s", err)
		return nil
	}
	return &MongoDB{
		Client:     mongoClient,
		Collection: mongoClient.Database(db).Collection(coll),
	}
}

// Store stores a message in MongoDB and returns its storage ID
func (mongo *MongoDB) Store(m *data.Message) (string, error) {
	_, err := mongo.Collection.InsertOne(context.Background(), m)
	if err != nil {
		log.Printf("Error inserting message: %s", err)
		return "", err
	}
	return string(m.ID), nil
}

// Count returns the number of stored messages
func (mongo *MongoDB) Count() int {
	c, _ := mongo.Collection.EstimatedDocumentCount(context.Background())
	return int(c)
}

// Search finds messages matching the query
func (mongo *MongoDB) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	messages := &data.Messages{}

	var field = "raw.data"
	switch kind {
	case "to":
		field = "raw.to"
	case "from":
		field = "raw.from"
	}

	data, err := mongo.Collection.Find(context.Background(), bson.M{field: primitive.Regex{Pattern: query, Options: "i"}}, options.Find().SetSkip(int64(start)).SetLimit(int64(limit)).SetSort(bson.D{{"created", 1}}).SetProjection(bson.M{
		"id":              1,
		"_id":             1,
		"from":            1,
		"to":              1,
		"content.headers": 1,
		"content.size":    1,
		"created":         1,
		"raw":             1,
	}))

	if data !=nil {
		data.All(context.Background(), messages)
	}

	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, 0, err
	}

	var count int64
	count,_ = mongo.Collection.CountDocuments(context.Background(), bson.M{field: primitive.Regex{Pattern: query, Options: "i"}}, options.Count())

	return messages, int(count), nil
}

// List returns a list of messages by index
func (mongo *MongoDB) List(start int, limit int) (*data.Messages, error) {
	messages := &data.Messages{}
	data, err := mongo.Collection.Find(context.Background(), bson.M{}, options.Find().SetSkip(int64(start)).SetLimit(int64(limit)).SetSort(bson.D{{"created", 1}}).SetProjection(bson.M{
		"id":              1,
		"_id":             1,
		"from":            1,
		"to":              1,
		"content.headers": 1,
		"content.size":    1,
		"created":         1,
		"raw":             1,
	}))

	if data != nil {
		data.All(context.Background(), messages)
	}

	if err != nil {
		log.Printf("Error loading messages: %s", err)
		return nil, err
	}
	return messages, nil
}

// DeleteOne deletes an individual message by storage ID
func (mongo *MongoDB) DeleteOne(id string) error {
	_, err := mongo.Collection.DeleteOne(context.Background(), bson.M{"id": id})
	return err
}

// DeleteAll deletes all messages stored in MongoDB
func (mongo *MongoDB) DeleteAll() error {
	_, err := mongo.Collection.DeleteMany(context.Background(), bson.M{})
	return err
}

// Load loads an individual message by storage ID
func (mongo *MongoDB) Load(id string) (*data.Message, error) {
	result := &data.Message{}
	err := mongo.Collection.FindOne(context.Background(), bson.M{"id": id}, options.FindOne()).Decode(&result)
	if err != nil {
		log.Printf("Error loading message: %s", err)
		return nil, err
	}
	return result, nil
}
