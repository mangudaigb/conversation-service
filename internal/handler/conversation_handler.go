package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangudaigb/conversation-service/internal/svc"
	"github.com/mangudaigb/conversation-service/pkg/dhauli"
	"github.com/mangudaigb/dhauli-base/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateType string

const (
	Query  UpdateType = "query"
	Answer UpdateType = "answer"
)

type ConversationRequest struct {
	UserID         string                 `json:"userId,omitempty" binding:"required"`
	WorkflowId     string                 `json:"workflowId,omitempty" binding:"required"`
	SessionId      string                 `json:"sessionId,omitempty" binding:"required"`
	ConversationId string                 `json:"conversationId,omitempty"`
	UpdateType     UpdateType             `json:"updateType,omitempty"`
	Data           dhauli.InteractionStub `json:"data,omitempty"`
}

type ConversationHandler struct {
	log  *logger.Logger
	svc  svc.ConversationService
	iSvc svc.InteractionService
}

func NewConversationHandler(log *logger.Logger, svc svc.ConversationService, iSvc svc.InteractionService) *ConversationHandler {
	return &ConversationHandler{
		log:  log,
		svc:  svc,
		iSvc: iSvc,
	}
}

func (ch *ConversationHandler) GetConversationsForUser(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	docs, err := ch.svc.GetConversationList(c.Request.Context(), uid)
	if err != nil {
		ch.log.Errorf("Error getting conversation list for user %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	}
	c.JSON(http.StatusOK, docs)
}

func (ch *ConversationHandler) GetConversationById(c *gin.Context) {
	id := c.Param("cid")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction ID is required"})
		return
	}

	doc, err := ch.svc.GetConversationById(c.Request.Context(), id)
	if err != nil {
		ch.log.Errorf("Error getting conversation %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
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

	if req.Data.Query == "" {
		ch.log.Errorf("Query is required to create a conversation")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
		return
	}
	cid := primitive.NewObjectID().Hex()
	inter, err := ch.iSvc.CreateInteraction(c.Request.Context(), &dhauli.Interaction{
		ID:             primitive.NewObjectID().Hex(),
		WorkflowID:     req.WorkflowId,
		SessionID:      req.SessionId,
		ConversationID: cid,
		Query:          req.Data.Query,
	})
	if err != nil {
		ch.log.Errorf("Error creating interaction: %v for conversations", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create interaction" + err.Error()})
		return
	}

	conversation := dhauli.Conversation{
		ID:         cid,
		WorkflowID: req.WorkflowId,
		SessionID:  req.SessionId,
		Interactions: []dhauli.InteractionStub{
			{
				ID:    inter.ID,
				Query: req.Data.Query,
			},
		},
	}
	createdConversation, err := ch.svc.CreateConversation(c.Request.Context(), &conversation)
	if err != nil {
		ch.log.Errorf("Error creating conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation" + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdConversation)
}

func (ch *ConversationHandler) UpdateConversation(c *gin.Context) {
	conversationId := c.Param("cid")
	if conversationId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}
	var req ConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ch.log.Errorf("Error parsing request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	var conversation *dhauli.Conversation
	var err error
	if req.UpdateType == Answer {
		if req.Data.ID == "" {
			ch.log.Errorf("Interaction ID is required")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction ID is required"})
			return
		}
		conversation, err = ch.svc.UpdateInteractionAnswer(c.Request.Context(), conversationId, req.Data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update answer in conversation: " + err.Error()})
			return
		}
	} else if req.UpdateType == Query {
		interactionStub := dhauli.InteractionStub{
			ID:    primitive.NewObjectID().Hex(),
			Query: req.Data.Query,
		}
		conversation, err = ch.svc.AddInteractionByConversationId(c.Request.Context(), conversationId, interactionStub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update answer in conversation: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, conversation)
}

func (ch *ConversationHandler) DeleteConversation(c *gin.Context) {
	conversationId := c.Param("cid")
	if conversationId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}
	err := ch.svc.DeleteConversation(c.Request.Context(), conversationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation: " + err.Error()})
		return
	}
}
