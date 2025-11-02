package internal

import (
	"context"
	"errors"

	"github.com/mangudaigb/conversation-memory/internal/consumer"
	"github.com/mangudaigb/dhauli-base/consumer/messaging"
	"github.com/mangudaigb/dhauli-base/logger"
)

type MessageHandler struct {
	log *logger.Logger
	ih  consumer.InteractionMsgHandler
	ch  consumer.ConversationMsgHandler
}

func (mh *MessageHandler) HandlerFunc(ctx context.Context, envelope *messaging.Envelope) *messaging.Envelope {
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
			return messaging.MessageError(envelope, 500, errors.New("interaction handler error"), false)
		}
		responseMsg, err := messaging.NewMessageFromOld(message, "interaction", message.Action, inter)
		if err != nil {
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
			return messaging.MessageError(envelope, 500, errors.New("interaction handler error"), false)
		}
		responseMsg, err := messaging.NewMessageFromOld(message, "interaction", message.Action, conv)
		if err != nil {
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

func NewMessageHandler(log *logger.Logger, ih consumer.InteractionMsgHandler, ch consumer.ConversationMsgHandler) *MessageHandler {
	return &MessageHandler{
		log: log,
		ih:  ih,
		ch:  ch,
	}
}
