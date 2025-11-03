package repo

import (
	"context"
	"errors"
	"time"

	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepository interface {
	GetByID(ctx context.Context, id string) (*dhauli.Conversation, error)
	Create(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	Update(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	Delete(ctx context.Context, id string) error
	Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.Conversation, error)
	Close()
}

type MongoConversationRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func (mcr *MongoConversationRepository) GetByID(ctx context.Context, id string) (*dhauli.Conversation, error) {
	conversationDoc := &dhauli.Conversation{}
	err := mcr.collection.FindOne(ctx, dhauli.Conversation{ID: id}).Decode(conversationDoc)
	if err != nil {
		mcr.log.Errorf("Error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return conversationDoc, nil
}

func (mcr *MongoConversationRepository) Create(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = now
	conversation.Version = 1
	result, err := mcr.collection.InsertOne(ctx, conversation)
	if err != nil {
		mcr.log.Errorf("Error inserting conversation: %v", err)
	}
	return mcr.GetByID(ctx, result.InsertedID.(string))
}

func (mcr *MongoConversationRepository) Update(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	filter := bson.M{
		"_id":     conversation.ID,
		"version": conversation.Version,
	}
	conversation.Version = conversation.Version + 1
	conversation.UpdatedAt = time.Now()
	update := bson.M{"$set": conversation}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedConversation dhauli.Conversation
	err := mcr.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(updatedConversation)
	if err != nil {
		mcr.log.Errorf("Error updating conversation: %v", err)
		return nil, err
	}
	return &updatedConversation, nil
}

func (mcr *MongoConversationRepository) Delete(ctx context.Context, id string) error {
	_, err := mcr.collection.DeleteOne(ctx, dhauli.Conversation{ID: id})
	if err != nil {
		mcr.log.Errorf("Error deleting conversation in mongo: %v", err)
		return err
	}
	return nil
}

func (mcr *MongoConversationRepository) Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.Conversation, error) {
	var list []*dhauli.Conversation
	cursor, err := mcr.collection.Find(ctx, filter)
	if err != nil {
		mcr.log.Errorf("Error getting conversation for filter: %v", err)
		return nil, err
	}
	if err = cursor.All(ctx, &list); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			mcr.log.Infof("No conversations found for filter: %v", filter)
			return []*dhauli.Conversation{}, nil
		}
		mcr.log.Errorf("Error decoding conversation: %v", err)
		return nil, err
	}
	return list, nil
}

func (mcr *MongoConversationRepository) Close() {
	err := mcr.collection.Database().Client().Disconnect(context.Background())
	if err != nil {
		mcr.log.Errorf("Error closing mongo client for conversation: %v", err)
	}
}

func NewConversationRepository(cfg *config.Config, log *logger.Logger, client mongo.Client, collection string) *MongoConversationRepository {
	col := client.Database(cfg.Mongo.Database).Collection(collection)
	return &MongoConversationRepository{
		log:        log,
		collection: col,
	}
}
