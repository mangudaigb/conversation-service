package contracts

import "github.com/mangudaigb/dhauli-base/consumer/messaging"

const (
	CreateConversation messaging.EventName = "CreateConversation"
	UpdateConversation messaging.EventName = "UpdateConversation"
	DeleteConversation messaging.EventName = "DeleteConversation"

	ConversationCreated messaging.EventName = "ConversationCreated"
	ConversationUpdated messaging.EventName = "ConversationUpdated"
	ConversationDeleted messaging.EventName = "ConversationDeleted"

	CreateInteraction  messaging.EventName = "CreateInteraction"
	UpdateInteraction  messaging.EventName = "UpdateInteraction"
	InteractionCreated messaging.EventName = "InteractionCreated"
	InteractionUpdated messaging.EventName = "InteractionUpdated"
)
