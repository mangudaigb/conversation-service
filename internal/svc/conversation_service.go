package svc

import (
	"context"

	"github.com/mangudaigb/conversation-service/internal/repo"
	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationService interface {
	CreateConversation(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	GetConversationById(ctx context.Context, cid string) (*dhauli.Conversation, error)
	GetConversationList(ctx context.Context, userId string) ([]*dhauli.Conversation, error)
	AddInteractionByConversationId(ctx context.Context, cid string, stub dhauli.InteractionStub) (*dhauli.Conversation, error)
	UpdateInteractionAnswer(ctx context.Context, cid string, stub dhauli.InteractionStub) (*dhauli.Conversation, error)
	DeleteConversation(ctx context.Context, cid string) error
}

type conversationService struct {
	log  *logger.Logger
	repo repo.ConversationRepository
	iSvc InteractionService
}

func (cs conversationService) GetConversationList(ctx context.Context, userId string) ([]*dhauli.Conversation, error) {
	convs, err := cs.repo.Filter(ctx, map[string]interface{}{
		"userId": userId,
	})
	if err != nil {
		cs.log.Errorf("Error getting conversation list for user: %s err: %v", userId, err)
		return nil, err
	}
	return convs, nil
}

// TODO Convert to single transaction
func (cs conversationService) CreateConversation(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	cid := primitive.NewObjectID().Hex()
	if len(conversation.Interactions) != 0 {
		inter := dhauli.Interaction{
			ID:             primitive.NewObjectID().Hex(),
			WorkflowID:     conversation.WorkflowID,
			SessionID:      conversation.SessionID,
			ConversationID: cid,
			Query:          conversation.Interactions[0].Query,
		}
		createdInteraction, err := cs.iSvc.CreateInteraction(ctx, &inter)
		if err != nil {
			cs.log.Errorf("Error creating interaction for conversation: %v", err)
			return nil, err
		}
		conversation.Interactions = []dhauli.InteractionStub{
			{
				ID:    createdInteraction.ID,
				Query: createdInteraction.Query,
			},
		}
	}
	return cs.repo.Create(ctx, conversation)
}

func (cs conversationService) GetConversationById(ctx context.Context, cid string) (*dhauli.Conversation, error) {
	c, err := cs.repo.GetByID(ctx, cid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", cid, err)
		return nil, err
	}
	return c, nil
}

func (cs conversationService) AddInteractionByConversationId(ctx context.Context, cid string, stub dhauli.InteractionStub) (*dhauli.Conversation, error) {
	c, err := cs.repo.GetByID(ctx, cid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", cid, err)
		return nil, err
	}
	for _, in := range c.Interactions {
		if in.ID == stub.ID {
			in.Answer = stub.Answer
			return cs.repo.Update(ctx, c)
		}
	}
	c.Interactions = append(c.Interactions, stub)
	return cs.repo.Update(ctx, c)
}

func (cs conversationService) UpdateInteractionAnswer(ctx context.Context, cid string, stub dhauli.InteractionStub) (*dhauli.Conversation, error) {
	c, err := cs.repo.GetByID(ctx, cid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", cid, err)
		return nil, err
	}
	for i, in := range c.Interactions {
		if in.ID == stub.ID {
			c.Interactions[i].Answer = stub.Answer
			break
		}
	}
	return cs.repo.Update(ctx, c)
}

func (cs conversationService) DeleteConversation(ctx context.Context, cid string) error {
	return cs.repo.Delete(ctx, cid)
}

func NewConversationService(log *logger.Logger, repo repo.ConversationRepository) ConversationService {
	return &conversationService{
		log:  log,
		repo: repo,
	}
}
