package dhauli

import "time"

type Memento struct {
	Index int    `json:"index"`
	State string `json:"state"`
	Actor string `json:"actor"`
}

type Conversation struct {
	ID                string    `json:"id" bson:"_id,omitempty"`
	WorkflowID        string    `json:"workflowId" bson:"name"`
	SessionID         string    `json:"sessionId" bson:"sessionId"`
	ConversationID    string    `json:"conversationId" bson:"conversationId"`
	Context           string    `json:"context" bson:"context"`
	Response          string    `json:"response" bson:"response"`
	CurrContextIndex  int       `json:"currContextIndex" bson:"currContextIndex"`
	CurrResponseIndex int       `json:"currResponseIndex" bson:"currResponseIndex"`
	ContextMemento    []Memento `json:"contexts" bson:"context_memento"`
	ResponsesMemento  []Memento `json:"responses" bson:"response_memento"`
	CreatedAt         time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt" bson:"updatedAt"`
}
