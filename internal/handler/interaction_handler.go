package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangudaigb/conversation-service/internal/svc"
	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/logger"
)

type Type string

const (
	CONTEXT Type = "context"
	QUERY   Type = "query"
	ANSWER  Type = "answer"
)

type InteractionRequest struct {
	WorkflowId     string `json:"workflowId,omitempty" binding:"required"`
	SessionId      string `json:"sessionId,omitempty" binding:"required"`
	ConversationId string `json:"conversationId,omitempty" binding:"required"`
	InteractionId  string `json:"interactionId,omitempty"`
	Actor          string `json:"actor,omitempty"`
	Action         string `json:"action,omitempty"`
	Type           Type   `json:"type"`
	Data           string `json:"data"`
	Version        int    `json:"version,omitempty"`
}

type InteractionHandler struct {
	log  *logger.Logger
	iSvc svc.InteractionService
}

func NewInteractionHandler(log *logger.Logger, svc svc.InteractionService) *InteractionHandler {
	return &InteractionHandler{
		log:  log,
		iSvc: svc,
	}
}

func (ch *InteractionHandler) GetInteractionsForConversation(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}
	interactions, err := ch.iSvc.GetInteractionByConversationId(c.Request.Context(), cid)
	if err != nil {
		ch.log.Errorf("Error getting interactions for conversation id %s: %v", cid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	}
	c.JSON(http.StatusOK, interactions)
}

func (ch *InteractionHandler) GetInteractionById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction ID is required"})
		return
	}

	doc, err := ch.iSvc.GetInteractionById(c.Request.Context(), id)
	if err != nil {
		ch.log.Errorf("Error getting conversation %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interaction not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (ch *InteractionHandler) CreateInteraction(c *gin.Context) {
	var req InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ch.log.Errorf("Error parsing request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
	interaction := dhauli.Interaction{
		WorkflowID:     req.WorkflowId,
		SessionID:      req.SessionId,
		ConversationID: req.ConversationId,
	}
	if req.Type == CONTEXT {
		interaction.Context = req.Data
	} else if req.Type == ANSWER {
		interaction.Answer = req.Data
	} else if req.Type == QUERY {
		interaction.Query = req.Data
	}
	createdInteraction, err := ch.iSvc.CreateInteraction(c.Request.Context(), &interaction)
	if err != nil {
		ch.log.Errorf("Error creating conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation" + err.Error()})
		return
	}
	//iStub := dhauli.InteractionStub{
	//	ID:     createdInteraction.ID,
	//	Query:  createdInteraction.Query,
	//	Answer: createdInteraction.Answer,
	//}
	//_, err = ch.cSvc.AddInteractionByConversationId(c.Request.Context(), createdInteraction.ConversationID, iStub)
	//if err != nil {
	//	ch.log.Errorf("Error adding interaction to conversation: %v", err)
	//}

	c.JSON(http.StatusCreated, createdInteraction)
}

func (ch *InteractionHandler) UpdateInteraction(c *gin.Context) {
	iid := c.Param("iid")
	if iid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction ID is required"})
		return
	}
	var req InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ch.log.Errorf("Error parsing request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
	if req.InteractionId == "" {
		ch.log.Errorf("Interaction ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction ID is required"})
	}
	var in *dhauli.Interaction
	var err error
	if req.Type == CONTEXT {
		in, err = ch.iSvc.UpdateContextInInteraction(c.Request.Context(), iid, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update context in interaction: " + err.Error()})
			return
		}
	} else if req.Type == ANSWER {
		in, err = ch.iSvc.UpdateAnswerInInteraction(c.Request.Context(), iid, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update answer in interaction: " + err.Error()})
			return
		}
	} else {
		in, err = ch.iSvc.UpdateQueryInInteraction(c.Request.Context(), iid, req.Data, req.Actor, req.Action, req.Version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update context in interaction: " + err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, in)
}
