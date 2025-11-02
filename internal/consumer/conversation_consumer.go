package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mangudaigb/conversation-memory/internal/handler"
	"github.com/mangudaigb/conversation-memory/internal/svc"
	"github.com/mangudaigb/conversation-memory/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/consumer/messaging"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationMsgHandler struct {
	log  *logger.Logger
	cSvc svc.ConversationService
}

func (cmh *ConversationMsgHandler) ConversationHandlerFunc(ctx context.Context, message messaging.Message, action messaging.Action) (*dhauli.Conversation, error) {
	var err error
	var out *dhauli.Conversation
	if action == "create" {
		out, err = cmh.handleCreate(ctx, message)
	} else if action == "update" {
		out, err = cmh.cSvc.UpdateInteractionAnswer(ctx, message.ID, dhauli.InteractionStub{})
	} else if action == "get" {
		out, err = cmh.cSvc.GetConversationById(ctx, message.ID)
	}
	return out, err
}

func (cmh *ConversationMsgHandler) handleCreate(ctx context.Context, msg messaging.Message) (*dhauli.Conversation, error) {
	var req handler.ConversationRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		cmh.log.Errorf("Error unmarshalling message for id: %s with %v", msg.ID, err)
		return nil, err
	}
	c := dhauli.Conversation{
		ID:         primitive.NewObjectID().Hex(),
		WorkflowID: req.WorkflowId,
		SessionID:  req.SessionId,
		UserID:     req.UserID,
	}
	createdConversation, err := cmh.cSvc.CreateConversation(ctx, &c)
	if err != nil {
		cmh.log.Errorf("Error creating conversation: %v", err)
		return nil, err
	}
	return createdConversation, nil
}

func (cmh *ConversationMsgHandler) handleUpdate(ctx context.Context, msg messaging.Message) (*dhauli.Conversation, error) {
	var req handler.ConversationRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		cmh.log.Errorf("Error unmarshalling message for id: %s with %v", msg.ID, err)
		return nil, err
	}

	if req.Data.ID == "" {
		cmh.log.Errorf("Interaction ID is required")
		return nil, errors.New("interaction ID is required to update conversation")
	}

	var conversation *dhauli.Conversation
	var err error
	if req.UpdateType == handler.Answer {
		conversation, err = cmh.cSvc.UpdateInteractionAnswer(ctx, req.ConversationId, req.Data)
	} else if req.UpdateType == handler.Query {
		interactionStub := dhauli.InteractionStub{
			ID:    primitive.NewObjectID().Hex(),
			Query: req.Data.Query,
		}
		conversation, err = cmh.cSvc.AddInteractionByConversationId(ctx, req.ConversationId, interactionStub)
	}
	if err != nil {
		cmh.log.Errorf("Error updating conversation: %v", err)
		return nil, err
	}

	return conversation, nil
}

func (cmh *ConversationMsgHandler) handleGet(ctx context.Context, msg messaging.Message) (*dhauli.Conversation, error) {
	var req handler.ConversationRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		cmh.log.Errorf("Error unmarshalling message for id: %s with %v", msg.ID, err)
		return nil, err
	}

	if req.ConversationId == "" {
		return nil, errors.New("conversation id cannot be empty")
	}
	return cmh.cSvc.GetConversationById(ctx, req.ConversationId)
}

//func createEnvelope(oldMsg *messaging.Message, oldEnv *messaging.Envelope, conversation *dhauli.Conversation) *messaging.Envelope {
//	responseMsg, err := messaging.NewMessageFromOld(oldMsg, "conversation", oldMsg.Action, conversation)
//	if err != nil {
//		return messaging.MessageError(oldEnv, 500, errors.New("error creating response message"), false)
//	}
//	responseEnv := messaging.NewEnvelope(
//		&responseMsg,
//		messaging.WithCorrelationId(oldEnv.CorrelationId),
//		messaging.WithTraceId(oldEnv.TraceId),
//		messaging.WithIdempotencyKey(oldEnv.IdempotencyKey),
//		messaging.WithKind(messaging.RESPONSE),
//		messaging.WithEventName("success"),
//	)
//	return &responseEnv
//}
