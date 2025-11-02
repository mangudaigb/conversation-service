package repo

import (
	"context"
	"errors"

	"github.com/mangudaigb/conversation-memory/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrSessionNotFound = errors.New("Session not found")
)

type InteractionRepository interface {
	GetById(ctx context.Context, id string) (*dhauli.Interaction, error)
	Create(ctx context.Context, conversation *dhauli.Interaction) (*dhauli.Interaction, error)
	Update(ctx context.Context, conversation *dhauli.Interaction) (*dhauli.Interaction, error)
	Delete(ctx context.Context, id string) error
	Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.Interaction, error)
	Close()
}

type MongoInteractionRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewMongoInteractionRepository(cfg *config.Config, log *logger.Logger, client mongo.Client, collection string) *MongoInteractionRepository {
	col := client.Database(cfg.Mongo.Database).Collection(collection)
	return &MongoInteractionRepository{
		log:        log,
		collection: col,
	}
}

func (msr *MongoInteractionRepository) GetById(ctx context.Context, id string) (*dhauli.Interaction, error) {
	conversationDoc := &dhauli.Interaction{}
	err := msr.collection.FindOne(ctx, dhauli.Interaction{ID: id}).Decode(conversationDoc)
	if err != nil {
		msr.log.Errorf("Error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return conversationDoc, nil
}

func (msr *MongoInteractionRepository) Create(ctx context.Context, conversation *dhauli.Interaction) (*dhauli.Interaction, error) {
	result, err := msr.collection.InsertOne(ctx, conversation)
	if err != nil {
		msr.log.Errorf("Error inserting conversation in mongo: %v", err)
		return nil, err
	}

	id := result.InsertedID.(string)
	var interactionDoc dhauli.Interaction
	err = msr.collection.FindOne(ctx, dhauli.Interaction{ID: id}).Decode(&interactionDoc)
	if err != nil {
		msr.log.Errorf("Was able to save the conversation but error getting conversation for id: %s err: %v", id, err)
		return nil, err
	}
	return &interactionDoc, nil
}

func (msr *MongoInteractionRepository) Update(ctx context.Context, interaction *dhauli.Interaction) (*dhauli.Interaction, error) {
	filter := bson.M{"_id": interaction.ID}
	update := bson.M{
		"$set": bson.M{
			"context":   interaction.Context,
			"query":     interaction.Query,
			"Answer":    interaction.Answer,
			"updatedAt": interaction.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedInteraction dhauli.Interaction
	err := msr.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedInteraction)
	if errors.Is(err, mongo.ErrNoDocuments) {
		msr.log.Errorf("Error updating interaction in mongo: %v", err)
		return nil, err
	}
	return &updatedInteraction, nil
}

func (msr *MongoInteractionRepository) Delete(ctx context.Context, id string) error {
	_, err := msr.collection.DeleteOne(ctx, dhauli.Interaction{ID: id})
	if err != nil {
		msr.log.Errorf("Error deleting conversation in mongo: %v", err)
		return err
	}
	return nil
}

func (msr *MongoInteractionRepository) Filter(ctx context.Context, filter map[string]interface{}) ([]*dhauli.Interaction, error) {
	var list []*dhauli.Interaction
	cursor, err := msr.collection.Find(ctx, filter)
	if err != nil {
		msr.log.Errorf("Error getting conversation for filter: %v", err)
		return nil, err
	}
	if err = cursor.All(ctx, &list); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			msr.log.Infof("No conversations found for filter: %v", filter)
			return []*dhauli.Interaction{}, nil
		}
		msr.log.Errorf("Error getting conversation for filter: %v", err)
		return nil, err
	}
	return list, nil
}

func (msr *MongoInteractionRepository) Close() {
	err := msr.collection.Database().Client().Disconnect(context.Background())
	if err != nil {
		msr.log.Errorf("Error closing mongo client for conversation: %v", err)
		return
	}
}
