package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type InteractionHistoryRepository interface {
	GetById(id string) (*dhauli.InteractionHistory, error)
	Create(ctx context.Context, conversation *dhauli.InteractionHistory) (*dhauli.InteractionHistory, error)
	Delete(ctx context.Context, id string)
	Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.InteractionHistory, error)
	Close()
}

type MongoInteractionHistoryRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewInteractionHistoryRepository(cfg *config.Config, log *logger.Logger, client mongo.Client, collection string) *MongoInteractionHistoryRepository {
	col := client.Database(cfg.Mongo.Database).Collection(collection)
	return &MongoInteractionHistoryRepository{
		log:        log,
		collection: col,
	}
}

func (msr MongoInteractionHistoryRepository) GetById(id string) (*dhauli.InteractionHistory, error) {
	interactionDoc := &dhauli.InteractionHistory{}
	filter := bson.M{"_id": id}
	err := msr.collection.FindOne(context.Background(), filter).Decode(interactionDoc)
	if err != nil {
		msr.log.Errorf("Error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return interactionDoc, nil
}

func (msr MongoInteractionHistoryRepository) Create(ctx context.Context, interactionHistory *dhauli.InteractionHistory) (*dhauli.InteractionHistory, error) {
	if interactionHistory.ID == "" {
		return nil, fmt.Errorf("interaction History id cannot be empty")
	}
	interactionHistory.CreatedAt = time.Now()
	ch, err := msr.collection.InsertOne(ctx, interactionHistory)
	if err != nil {
		msr.log.Errorf("Error inserting conversation history in mongo: %v", err)
		return nil, err
	}
	return msr.GetById(ch.InsertedID.(string))
}

func (msr MongoInteractionHistoryRepository) Delete(ctx context.Context, id string) {
	_, err := msr.collection.DeleteOne(ctx, dhauli.InteractionHistory{ID: id})
	if err != nil {
		msr.log.Errorf("Error deleting conversation in mongo: %v", err)
		return
	}
}

func (msr MongoInteractionHistoryRepository) Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.InteractionHistory, error) {
	cursor, err := msr.collection.Find(ctx, filter)
	if err != nil {
		msr.log.Errorf("Error finding interaction history: %v", err)
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			msr.log.Errorf("Error closing context history cursor: %v", closeErr)
		}
	}()
	var interactionHistory []*dhauli.InteractionHistory
	if err = cursor.All(ctx, &interactionHistory); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []*dhauli.InteractionHistory{}, nil
		}
		msr.log.Errorf("Error decoding interaction history: %v", err)
		return nil, err
	}
	return interactionHistory, nil

}

func (msr MongoInteractionHistoryRepository) Close() {
	err := msr.collection.Database().Client().Disconnect(context.Background())
	if err != nil {
		msr.log.Errorf("Error closing mongo client for conversation: %v", err)
		return
	}
}
