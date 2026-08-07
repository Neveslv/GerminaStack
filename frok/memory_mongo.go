package frok

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const memoryLimit int64 = 8

type MongoMemoryStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type memoryDocument struct {
	UserID    int64     `bson:"user_id"`
	Prompt    string    `bson:"prompt"`
	Reply     string    `bson:"reply"`
	CreatedAt time.Time `bson:"created_at"`
}

func NewMongoMemoryStore(ctx context.Context, uri, database string) (*MongoMemoryStore, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	collection := client.Database(database).Collection("frok_memories")
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return &MongoMemoryStore{client: client, collection: collection}, nil
}

func (s *MongoMemoryStore) Recall(ctx context.Context, userID int64) ([]Memory, error) {
	cursor, err := s.collection.Find(ctx, bson.D{{Key: "user_id", Value: userID}}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(memoryLimit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []memoryDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	memories := make([]Memory, len(documents))
	for index, document := range documents {
		memories[len(documents)-1-index] = Memory{Prompt: document.Prompt, Reply: document.Reply}
	}
	return memories, nil
}

func (s *MongoMemoryStore) Remember(ctx context.Context, userID int64, prompt, reply string) error {
	_, err := s.collection.InsertOne(ctx, memoryDocument{UserID: userID, Prompt: prompt, Reply: reply, CreatedAt: time.Now().UTC()})
	return err
}

func (s *MongoMemoryStore) Close() {
	_ = s.client.Disconnect(context.Background())
}
