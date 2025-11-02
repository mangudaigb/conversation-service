package svc

import (
	"context"
	"time"

	"github.com/mangudaigb/conversation-memory/internal/repo"
	"github.com/mangudaigb/conversation-memory/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InteractionHistoryService interface {
	GetInteractionHistoryById(id string) (*dhauli.InteractionHistory, error)
	AddHistoryForInteraction(ctx context.Context, interaction *dhauli.Interaction, actor string, action string) (*dhauli.InteractionHistory, error)
	GetHistoryForInteractionId(ctx context.Context, iid string) ([]*dhauli.InteractionHistory, error)
}

type interactionHistoryService struct {
	log                          *logger.Logger
	interactionHistoryRepository repo.InteractionHistoryRepository
}

func (i interactionHistoryService) GetInteractionHistoryById(id string) (*dhauli.InteractionHistory, error) {
	return i.interactionHistoryRepository.GetById(id)
}

func (i interactionHistoryService) AddHistoryForInteraction(ctx context.Context, interaction *dhauli.Interaction, actor string, action string) (*dhauli.InteractionHistory, error) {
	ih := &dhauli.InteractionHistory{
		ID:             primitive.NewObjectID().Hex(),
		WorkflowID:     interaction.WorkflowID,
		SessionID:      interaction.SessionID,
		ConversationID: interaction.ConversationID,
		InteractionID:  interaction.ID,
		Action:         action,
		Actor:          actor,
		Context:        interaction.Context,
		Query:          interaction.Query,
		Answer:         interaction.Answer,
		CreatedAt:      time.Now(),
	}
	ihDoc, err := i.interactionHistoryRepository.Create(ctx, ih)
	if err != nil {
		i.log.Errorf("Error creating interaction history: %v for id: %s", err, interaction.ID)
		return nil, err
	}
	return ihDoc, nil
}

func (i interactionHistoryService) GetHistoryForInteractionId(ctx context.Context, iid string) ([]*dhauli.InteractionHistory, error) {
	filter := bson.M{"interactionId": iid}
	list, err := i.interactionHistoryRepository.Filter(ctx, filter)
	if err != nil {
		i.log.Errorf("Error getting interaction history for interaction id: %s err: %v", iid, err)
		return nil, err
	}
	return list, nil
}

func NewInteractionHistoryService(log *logger.Logger, repo repo.InteractionHistoryRepository) InteractionHistoryService {
	return &interactionHistoryService{
		log:                          log,
		interactionHistoryRepository: repo,
	}
}
