package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mangudaigb/conversation-service/internal/handler"
	"github.com/mangudaigb/conversation-service/internal/svc"
	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/consumer/messaging"
	"github.com/mangudaigb/dhauli-base/logger"
)

type InteractionMsgHandler struct {
	log  *logger.Logger
	iSvc svc.InteractionService
}

func (ih *InteractionMsgHandler) InteractionHandlerFunc(ctx context.Context, message messaging.Message, action messaging.Action) (*dhauli.Interaction, error) {
	var err error
	var out *dhauli.Interaction
	if action == "create" {
		out, err = ih.handleCreate(ctx, message)
	} else if action == "update" {
		out, err = ih.handleUpdate(ctx, message)
	} else if action == "get" {
		out, err = ih.handleGet(ctx, message)
	}
	return out, err
}

func (ih *InteractionMsgHandler) handleCreate(ctx context.Context, msg messaging.Message) (*dhauli.Interaction, error) {
	var req handler.InteractionRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		ih.log.Errorf("Error unmarshalling message for id: %s with %v", msg.ID, err)
		return nil, err
	}
	interaction := dhauli.Interaction{
		WorkflowID:     req.WorkflowId,
		SessionID:      req.SessionId,
		ConversationID: req.ConversationId,
	}
	if req.Type == handler.CONTEXT {
		interaction.Context = req.Data
	} else if req.Type == handler.ANSWER {
		interaction.Answer = req.Data
	} else if req.Type == handler.QUERY {
		interaction.Query = req.Data
	}
	createdInteraction, err := ih.iSvc.CreateInteraction(ctx, &interaction)
	if err != nil {
		ih.log.Errorf("Error creating conversation: %v", err)
		return nil, err
	}
	return createdInteraction, nil
}

func (ih *InteractionMsgHandler) handleGet(ctx context.Context, msg messaging.Message) (*dhauli.Interaction, error) {
	var req handler.InteractionRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		ih.log.Errorf("Error unmarshalling message for id: %s with %v", msg.ID, err)
		return nil, err
	}
	if req.InteractionId == "" {
		return nil, errors.New("interaction id cannot be empty")
	}
	return ih.iSvc.GetInteractionById(ctx, req.InteractionId)
}

func (ih *InteractionMsgHandler) handleUpdate(ctx context.Context, msg messaging.Message) (*dhauli.Interaction, error) {
	var req handler.InteractionRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		ih.log.Errorf("Error parsing Data of message: %v", err)
		return nil, err
	}
	if req.InteractionId == "" {
		ih.log.Errorf("Interaction ID is required")
		return nil, errors.New("interaction id cannot be empty")
	}
	var in *dhauli.Interaction
	var err error
	if req.Type == handler.CONTEXT {
		in, err = ih.iSvc.UpdateContextInInteraction(ctx, req.InteractionId, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			ih.log.Errorf("Failed to update context in interaction: %v", err)
			return nil, err
		}
	} else if req.Type == handler.ANSWER {
		in, err = ih.iSvc.UpdateAnswerInInteraction(ctx, req.InteractionId, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			ih.log.Errorf("Failed to update answer in interaction: %v", err)
			return nil, err
		}
	} else {
		in, err = ih.iSvc.UpdateContextInInteraction(ctx, req.InteractionId, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			ih.log.Errorf("Failed to update context in interaction: %%v", err)
			return nil, err
		}
	}
	return in, nil
}

func NewInteractionMsgHandler(log *logger.Logger, iSvc svc.InteractionService) *InteractionMsgHandler {
	return &InteractionMsgHandler{
		log:  log,
		iSvc: iSvc,
	}
}
