package internal

import (
	"context"
	"errors"

	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/logger"
	"github.com/mangudaigb/short-memory/pkg/dhauli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrSessionNotFound = errors.New("Session not found")
)

type ConversationRepository interface {
	GetById(ctx context.Context, id string) (*dhauli.Conversation, error)
	Create(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	Update(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	Delete(ctx context.Context, id string) error
	Close()
}

type MongoConversationRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewMongoConversationRepository(cfg *config.Config, log *logger.Logger, client mongo.Client, collection string) *MongoConversationRepository {
	col := client.Database(cfg.Mongo.Database).Collection(collection)
	return &MongoConversationRepository{
		log:        log,
		collection: col,
	}
}

func (msr *MongoConversationRepository) GetById(ctx context.Context, id string) (*dhauli.Conversation, error) {
	conversationDoc := &dhauli.Conversation{}
	err := msr.collection.FindOne(ctx, dhauli.Conversation{ID: id}).Decode(conversationDoc)
	if err != nil {
		msr.log.Errorf("Error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return conversationDoc, nil
}

func (msr *MongoConversationRepository) Create(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	result, err := msr.collection.InsertOne(ctx, conversation)
	if err != nil {
		msr.log.Errorf("Error inserting conversation in mongo: %v", err)
		return nil, err
	}

	id := result.InsertedID.(string)
	var conversationDoc dhauli.Conversation
	err = msr.collection.FindOne(ctx, dhauli.Conversation{ID: id}).Decode(&conversationDoc)
	if err != nil {
		msr.log.Errorf("Was able to save the conversation but error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return &conversationDoc, nil
}

func (msr *MongoConversationRepository) Update(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	filter := bson.M{"_id": conversation.ID}
	update := bson.M{
		"$set": bson.M{
			"context":           conversation.Context,
			"response":          conversation.Response,
			"currContextIndex":  conversation.CurrContextIndex,
			"currResponseIndex": conversation.CurrResponseIndex,
			"context_memento":   conversation.ContextMemento,
			"response_memento":  conversation.ResponsesMemento,
			"updatedAt":         conversation.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedConversation dhauli.Conversation
	err := msr.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedConversation)
	if errors.Is(err, mongo.ErrNoDocuments) {
		msr.log.Errorf("Error updating conversation in mongo: %v", err)
		return nil, err
	}
	return &updatedConversation, nil
}

func (msr *MongoConversationRepository) Delete(ctx context.Context, id string) error {
	_, err := msr.collection.DeleteOne(ctx, dhauli.Conversation{ID: id})
	if err != nil {
		msr.log.Errorf("Error deleting conversation in mongo: %v", err)
		return err
	}
	return nil
}

func (msr *MongoConversationRepository) Close() {
	err := msr.collection.Database().Client().Disconnect(context.Background())
	if err != nil {
		msr.log.Errorf("Error closing mongo client for conversation: %v", err)
		return
	}
}
