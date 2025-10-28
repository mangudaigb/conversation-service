package internal

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mangudaigb/dhauli-base/logger"
	"github.com/mangudaigb/short-memory/pkg/dhauli"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Type string

const (
	CONTEXT  Type = "context"
	RESPONSE Type = "response"
)

type ConversationRequest struct {
	ConversationId string `json:"conversationId,omitempty"`
	WorkflowId     string `json:"workflowId"`
	SessionId      string `json:"sessionId"`
	Type           Type   `json:"type"`
	Data           string `json:"data"`
}

type ConversationHandler struct {
	log *logger.Logger
	svc ConversationService
}

func NewConversationHandler(log *logger.Logger, svc ConversationService) *ConversationHandler {
	return &ConversationHandler{
		log: log,
		svc: svc,
	}
}

func (ch *ConversationHandler) GetConversationById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	doc, err := ch.svc.GetConversationById(c.Request.Context(), id)
	if err != nil {
		ch.log.Errorf("Error getting conversation %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (ch *ConversationHandler) CreateConversation(c *gin.Context) {
	var req ConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ch.log.Errorf("Error parsing request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
	now := time.Now()
	var cid string
	if req.ConversationId != "" {
		cid = req.ConversationId
	} else {
		cid = primitive.NewObjectID().Hex()
	}
	cnv := dhauli.Conversation{
		ID:                primitive.NewObjectID().Hex(),
		WorkflowID:        req.WorkflowId,
		SessionID:         req.SessionId,
		ConversationID:    cid,
		CurrContextIndex:  0,
		CurrResponseIndex: 0,
		ContextMemento:    []dhauli.Memento{},
		ResponsesMemento:  []dhauli.Memento{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if req.Type == CONTEXT {
		cnv.Context = req.Data
	} else if req.Type == RESPONSE {
		cnv.Response = req.Data
	}
	conversation, err := ch.svc.CreateConversation(c.Request.Context(), &cnv)
	if err != nil {
		ch.log.Errorf("Error creating conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation" + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, conversation)
}

func (ch *ConversationHandler) UpdateConversation(c *gin.Context) {

}
