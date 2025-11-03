package cmd

import (
	"context"
	"fmt"

	"github.com/mangudaigb/conversation-service/internal"
	consumer2 "github.com/mangudaigb/conversation-service/internal/consumer"
	"github.com/mangudaigb/conversation-service/internal/repo"
	"github.com/mangudaigb/conversation-service/internal/svc"
	"github.com/mangudaigb/conversation-service/pkg"
	"github.com/mangudaigb/dhauli-base/config"
	"github.com/mangudaigb/dhauli-base/consumer"
	"github.com/mangudaigb/dhauli-base/db"
	"github.com/mangudaigb/dhauli-base/discover"
	"github.com/mangudaigb/dhauli-base/logger"
	"github.com/mangudaigb/dhauli-base/tracing"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println("Error reading the config file", err)
		panic(err)
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Println("Error creating logger", err)
		panic(err)
	}

	tp := tracing.InitTracerProvider(cfg, log)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Errorf("Error shutting down tracer provider: %v", err)
		}
	}()
	tr := tp.Tracer("conversation-service")

	registry := discover.NewRegistryInfo(cfg, log)
	registry.Register(discover.SERVICE)

	StartConsumer(context.Background(), cfg, tr, log)

	server := pkg.NewConversationServer(cfg, tr, log)
	server.Start()
}

func StartConsumer(ctx context.Context, cfg *config.Config, tr trace.Tracer, log *logger.Logger) {
	mongoClient, err := db.NewMongoClient(cfg, log)
	if err != nil {
		log.Fatalf("Error creating mongo client: %v", err)
	}

	var interactionHistoryRepo = repo.NewInteractionHistoryRepository(cfg, log, *mongoClient.Client, "interactions_history")
	var interactionRepo = repo.NewMongoInteractionRepository(cfg, log, *mongoClient.Client, "interactions")
	var conversationRepo = repo.NewConversationRepository(cfg, log, *mongoClient.Client, "conversations")
	var conversationSvc = svc.NewConversationService(log, conversationRepo)
	var interactionHistorySvc = svc.NewInteractionHistoryService(log, interactionHistoryRepo)
	var interactionSvc = svc.NewInteractionService(log, interactionRepo, interactionHistorySvc, conversationSvc)
	var interactionMsgHandler = consumer2.NewInteractionMsgHandler(log, interactionSvc)
	var conversationMsgHandler = consumer2.NewConversationMsgHandler(log, conversationSvc)

	var msgHandler = internal.NewMessageHandler(tr, log, interactionMsgHandler, conversationMsgHandler)

	log.Infof("Starting kafka consumer")
	csmr := consumer.NewKafkaConsumer(cfg, tr, log, msgHandler.HandlerFunc)
	defer csmr.Stop()

	go func() {
		err := csmr.Consume(context.Background())
		if err != nil { // Should never happen
			log.Errorf("Error running the consumer: %v", err)
			return
		}
	}()
}
