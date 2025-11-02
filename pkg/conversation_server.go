package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mangudaigb/conversation-memory/internal/handler"
	"github.com/mangudaigb/conversation-memory/internal/repo"
	"github.com/mangudaigb/conversation-memory/internal/svc"
	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/consumer"
	"github.com/mangudaigb/dhauli-base/db"
	"github.com/mangudaigb/dhauli-base/logger"
	"golang.org/x/net/context"
)

type HttpServer struct {
	*consumer.KafkaConsumer
	zkClient *db.ZkClient
}

type ConversationServer struct {
	log *logger.Logger
	cfg *config.Config
}

func NewConversationServer(cfg *config.Config, log *logger.Logger) *ConversationServer {
	return &ConversationServer{
		log: log,
		cfg: cfg,
	}
}

func SetupRouter(log *logger.Logger, iSvc svc.InteractionService, cSvc svc.ConversationService) *gin.Engine {
	r := gin.Default()
	interactionHandler := handler.NewInteractionHandler(log, iSvc)
	conversationHandler := handler.NewConversationHandler(log, cSvc, iSvc)

	routes := r.Group("/conversations")
	{
		routes.GET("", conversationHandler.GetConversationsForUser)
		routes.GET("/:cid", conversationHandler.GetConversationById)
		routes.POST("/", conversationHandler.CreateConversation)
		routes.PATCH("/:cid", conversationHandler.UpdateConversation)
		routes.DELETE("/:cid", conversationHandler.DeleteConversation)

		interactionRoutes := routes.Group("/:cid/interactions")
		{
			interactionRoutes.GET("", interactionHandler.GetInteractionsForConversation)
			interactionRoutes.GET("/:iid", interactionHandler.GetInteractionById)
			interactionRoutes.POST("/", interactionHandler.CreateInteraction)
			interactionRoutes.PATCH("/:iid", interactionHandler.UpdateInteraction)
		}
	}

	return r
}

func (s *ConversationServer) Start() {
	mongoClient, err := db.NewMongoClient(s.cfg, s.log)
	if err != nil {
		s.log.Fatalf("Error creating mongo client: %v", err)
	}
	var interactionHistoryRepo = repo.NewInteractionHistoryRepository(s.cfg, s.log, *mongoClient.Client, "interactions_history")
	var interactionRepo = repo.NewMongoInteractionRepository(s.cfg, s.log, *mongoClient.Client, "interactions")
	var conversationRepo = repo.NewConversationRepository(s.cfg, s.log, *mongoClient.Client, "conversations")
	var conversationSvc = svc.NewConversationService(s.log, conversationRepo)
	var interactionHistorySvc = svc.NewInteractionHistoryService(s.log, interactionHistoryRepo)
	var interactionSvc = svc.NewInteractionService(s.log, interactionRepo, interactionHistorySvc, conversationSvc)

	router := SetupRouter(s.log, interactionSvc, conversationSvc)

	serverAddr := fmt.Sprintf(":%d", s.cfg.Server.Port)

	server := &http.Server{
		Addr:           serverAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalf("Error starting server: %v", err)
		}
	}()
	s.log.Infof("Server listening on %s", serverAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	s.log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		s.log.Fatalf("Server forced to shutdown (timeout/error): %v", err)
	}
	s.log.Info("Server successfully exited.")
}
