package internal

import (
	"context"
	"time"

	"github.com/mangudaigb/dhauli-base/logger"
	"github.com/mangudaigb/short-memory/pkg/dhauli"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationService interface {
	CreateConversation(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error)
	GetConversationById(ctx context.Context, id string) (*dhauli.Conversation, error)
	UpdateContextInConversation(ctx context.Context, cid string, context string, actor string) (*dhauli.Conversation, error)
	UpdateResponseInConversation(ctx context.Context, cid string, response string, actor string) (*dhauli.Conversation, error)
	DeleteConversation(ctx context.Context, id string) error
}

type conversationService struct {
	log                    *logger.Logger
	conversationRepository ConversationRepository
}

func NewConversationService(log *logger.Logger, repo ConversationRepository) ConversationService {
	return &conversationService{
		log:                    log,
		conversationRepository: repo,
	}
}

func (cs conversationService) CreateConversation(ctx context.Context, conversation *dhauli.Conversation) (*dhauli.Conversation, error) {
	conversation.ID = primitive.NewObjectID().Hex()
	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = now
	conversation.CurrContextIndex = 0
	conversation.CurrResponseIndex = 0
	return cs.conversationRepository.Create(ctx, conversation)
}

func (cs conversationService) GetConversationById(ctx context.Context, id string) (*dhauli.Conversation, error) {
	return cs.conversationRepository.GetById(ctx, id)
}

func (cs conversationService) UpdateContextInConversation(ctx context.Context, cid string, context string, actor string) (*dhauli.Conversation, error) {
	con, err := cs.conversationRepository.GetById(ctx, cid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", cid, err)
		return nil, err
	}
	contextMemento := dhauli.Memento{
		Index: con.CurrContextIndex,
		State: context,
		Actor: actor,
	}
	con.ContextMemento = append(con.ContextMemento, contextMemento)
	con.Context = context
	con.CurrContextIndex++
	con.UpdatedAt = time.Now()
	return cs.conversationRepository.Update(ctx, con)
}

func (cs conversationService) UpdateResponseInConversation(ctx context.Context, cid string, response string, actor string) (*dhauli.Conversation, error) {
	con, err := cs.conversationRepository.GetById(ctx, cid)
	if err != nil {
		cs.log.Errorf("Error getting conversation for id: %s err: %v", cid, err)
		return nil, err
	}
	responseMemento := dhauli.Memento{
		Index: con.CurrContextIndex,
		State: response,
		Actor: actor,
	}
	con.ResponsesMemento = append(con.ResponsesMemento, responseMemento)
	con.Response = response
	con.CurrResponseIndex++
	con.UpdatedAt = time.Now()
	return cs.conversationRepository.Update(ctx, con)
}

func (cs conversationService) DeleteConversation(ctx context.Context, id string) error {
	return cs.conversationRepository.Delete(ctx, id)
}
