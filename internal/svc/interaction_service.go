package svc

import (
	"context"
	"errors"
	"time"

	"github.com/mangudaigb/conversation-memory/internal/repo"
	"github.com/mangudaigb/conversation-memory/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InteractionService interface {
	CreateInteraction(ctx context.Context, interaction *dhauli.Interaction) (*dhauli.Interaction, error)
	GetInteractionById(ctx context.Context, iid string) (*dhauli.Interaction, error)
	GetInteractionByConversationId(ctx context.Context, cid string) ([]*dhauli.Interaction, error)
	UpdateContextInInteraction(ctx context.Context, iid string, context string, actor, action string, version int) (*dhauli.Interaction, error)
	UpdateQueryInInteraction(ctx context.Context, iid string, context string, actor, action string, version int) (*dhauli.Interaction, error)
	UpdateAnswerInInteraction(ctx context.Context, iid string, response string, actor, action string, version int) (*dhauli.Interaction, error)
	DeleteInteraction(ctx context.Context, iid string) error
}

type interactionService struct {
	log                   *logger.Logger
	interactionRepository repo.InteractionRepository
	historySvc            InteractionHistoryService
	conversationSvc       ConversationService
}

func NewInteractionService(log *logger.Logger, repo repo.InteractionRepository, hSvc InteractionHistoryService, cSvc ConversationService) InteractionService {
	return &interactionService{
		log:                   log,
		interactionRepository: repo,
		historySvc:            hSvc,
		conversationSvc:       cSvc,
	}
}

// TODO Convert this to a single transaction
func (cs interactionService) CreateInteraction(ctx context.Context, interaction *dhauli.Interaction) (*dhauli.Interaction, error) {
	interaction.ID = primitive.NewObjectID().Hex()
	now := time.Now()
	interaction.CreatedAt = now
	interaction.UpdatedAt = now
	interaction.Version = 1
	cid := interaction.ConversationID
	if interaction.Query == "" {
		_, _ = cs.conversationSvc.AddInteractionByConversationId(ctx, cid, dhauli.InteractionStub{
			ID:    interaction.ID,
			Query: interaction.Query,
		})
	} else {
		_, _ = cs.conversationSvc.AddInteractionByConversationId(ctx, cid, dhauli.InteractionStub{
			ID:     interaction.ID,
			Query:  interaction.Query,
			Answer: interaction.Answer,
		})
	}
	return cs.interactionRepository.Create(ctx, interaction)
}

func (cs interactionService) GetInteractionById(ctx context.Context, id string) (*dhauli.Interaction, error) {
	return cs.interactionRepository.GetById(ctx, id)
}

func (cs interactionService) GetInteractionByConversationId(ctx context.Context, cid string) ([]*dhauli.Interaction, error) {
	filter := bson.M{"conversationId": cid}
	return cs.interactionRepository.Filter(ctx, filter)
}

func (cs interactionService) UpdateContextInInteraction(ctx context.Context, iid, context, actor, action string, version int) (*dhauli.Interaction, error) {
	interaction, err := cs.interactionRepository.GetById(ctx, iid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", iid, err)
		return nil, err
	}
	if interaction.Version != version {
		cs.log.Errorf("Error updating context for interaction %s. Version mismatch. Expected: %d, Actual: %d", iid, version, interaction.Version)
		return nil, errors.New("interaction version mismatch")
	}
	_, err = cs.historySvc.AddHistoryForInteraction(ctx, interaction, actor, action)
	if err != nil {
		cs.log.Errorf("Error adding history for interaction while updating context: %v", err)
		return nil, err
	}
	interaction.Context = context
	interaction.UpdatedAt = time.Now()
	return cs.interactionRepository.Update(ctx, interaction)
}

func (cs interactionService) UpdateQueryInInteraction(ctx context.Context, iid, query, actor, action string, version int) (*dhauli.Interaction, error) {
	interaction, err := cs.interactionRepository.GetById(ctx, iid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", iid, err)
		return nil, err
	}
	if interaction.Version != version {
		cs.log.Errorf("Error updating query for interaction %s. Version mismatch. Expected: %d, Actual: %d", iid, version, interaction.Version)
		return nil, errors.New("interaction version mismatch")
	}
	_, err = cs.historySvc.AddHistoryForInteraction(ctx, interaction, actor, action)
	if err != nil {
		cs.log.Errorf("Error adding history for interaction while updating query: %v", err)
		return nil, err
	}
	interaction.Query = query
	interaction.UpdatedAt = time.Now()
	return cs.interactionRepository.Update(ctx, interaction)
}

func (cs interactionService) UpdateAnswerInInteraction(ctx context.Context, iid, response, actor, action string, version int) (*dhauli.Interaction, error) {
	interaction, err := cs.interactionRepository.GetById(ctx, iid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", iid, err)
		return nil, err
	}
	if interaction.Version != version {
		cs.log.Errorf("Error updating answer for interaction %s. Version mismatch. Expected: %d, Actual: %d", iid, version, interaction.Version)
		return nil, errors.New("interaction version mismatch")
	}
	_, err = cs.historySvc.AddHistoryForInteraction(ctx, interaction, actor, action)
	if err != nil {
		cs.log.Errorf("Error adding history for interaction while updating answer: %v", err)
		return nil, err
	}
	interaction.Answer = response
	interaction.UpdatedAt = time.Now()
	return cs.interactionRepository.Update(ctx, interaction)
}

func (cs interactionService) DeleteInteraction(ctx context.Context, id string) error {
	return cs.interactionRepository.Delete(ctx, id)
}
