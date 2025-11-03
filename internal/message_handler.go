package internal

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/mangudaigb/conversation-service/internal/consumer"
	"github.com/mangudaigb/dhauli-base/consumer/messaging"
	"github.com/mangudaigb/dhauli-base/logger"
)

type MessageHandler struct {
	tr  trace.Tracer
	log *logger.Logger
	ih  *consumer.InteractionMsgHandler
	ch  *consumer.ConversationMsgHandler
}

func (mh *MessageHandler) HandlerFunc(ctx context.Context, envelope *messaging.Envelope) *messaging.Envelope {
	_, span := mh.tr.Start(ctx, "Generic Handler")
	defer span.End()

	kind := envelope.Kind
	//event := envelope.EventName
	message := envelope.Message
	mType := message.Type
	mAction := message.Action

	if kind != messaging.REQUEST {
		return messaging.EnvelopeError(*envelope, "kind is not of type request", true)
	}
	if mType == "interaction" {
		inter, err := mh.ih.InteractionHandlerFunc(ctx, message, mAction)
		if err != nil {
			mh.log.Errorf("Error handling interaction: %v", err)
			return messaging.MessageError(envelope, 500, errors.New("interaction handler error"), false)
		}
		responseMsg, err := messaging.NewMessageFromOld(message, "interaction", message.Action, inter)
		if err != nil {
			mh.log.Errorf("Error creating response message for interaction: %v", err)
			return messaging.MessageError(envelope, 500, errors.New("error creating response message"), false)
		}
		responseEnv := messaging.NewEnvelope(
			responseMsg,
			messaging.WithCorrelationId(envelope.CorrelationId),
			messaging.WithTraceId(envelope.TraceId),
			messaging.WithIdempotencyKey(envelope.IdempotencyKey),
			messaging.WithKind(messaging.RESPONSE),
			messaging.WithEventName("success"),
		)
		return &responseEnv
	} else if mType == "conversation" {
		conv, err := mh.ch.ConversationHandlerFunc(ctx, message, mAction)
		if err != nil {
			mh.log.Errorf("Error handling conversation: %v", err)
			return messaging.MessageError(envelope, 500, errors.New("interaction handler error"), false)
		}
		responseMsg, err := messaging.NewMessageFromOld(message, "interaction", message.Action, conv)
		if err != nil {
			mh.log.Errorf("Error creating response message for conversation: %v", err)
			return messaging.MessageError(envelope, 500, errors.New("error creating response message"), false)
		}
		responseEnv := messaging.NewEnvelope(
			responseMsg,
			messaging.WithCorrelationId(envelope.CorrelationId),
			messaging.WithTraceId(envelope.TraceId),
			messaging.WithIdempotencyKey(envelope.IdempotencyKey),
			messaging.WithKind(messaging.RESPONSE),
			messaging.WithEventName("success"),
		)
		return &responseEnv
	}
	return messaging.MessageError(envelope, 500, errors.New("message type is not correct for handler"), true)
}

func NewMessageHandler(tr trace.Tracer, log *logger.Logger, ih *consumer.InteractionMsgHandler, ch *consumer.ConversationMsgHandler) *MessageHandler {
	return &MessageHandler{
		tr:  tr,
		log: log,
		ih:  ih,
		ch:  ch,
	}
}
