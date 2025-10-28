package internal

import (
	"context"
	"encoding/json"

	"github.com/mangudaigb/dhauli-base/logger"
	messages "github.com/mangudaigb/dhauli-base/types/message"
	"github.com/mangudaigb/short-memory/pkg/dhauli"
	"github.com/segmentio/kafka-go"
)

type MsgHandler struct {
	log *logger.Logger
	svc ConversationService
}

func (mh *MsgHandler) MsgHandlerFunc(ctx context.Context, log *logger.Logger, msg kafka.Message) (*messages.MessageResponse, error) {
	var incomingMsg messages.MessageRequest
	if err := json.Unmarshal(msg.Value, &incomingMsg); err != nil {
		mh.log.Errorf("Error unmarshalling message: %v", err)
		return nil, err
	}
	action := incomingMsg.Payload.Action
	data := incomingMsg.Payload.Data

	if action == "create" {

	}

	return nil, nil
}

func (mh *MsgHandler) handleCreate(ctx context.Context, incomingMsg messages.MessageRequest) *messages.MessageResponse {
	conv, err := mh.PDataToC(incomingMsg.Payload.Data)
	if err != nil {
		mh.log.Errorf("Error converting message payload to conversation entity for request id: %s with error : %v", incomingMsg.Id, err)
		return &messages.MessageResponse{
			Workflow: incomingMsg.Workflow,
			Status:   429,
			ErrorMsg: err.Error(),
		}
	}
	response, err := mh.svc.CreateConversation(ctx, conv)
	if err != nil {
		return &messages.MessageResponse{
			Workflow: incomingMsg.Workflow,
			Status:   503,
			ErrorMsg: err.Error(),
		}
	}
	dataMap, err := mh.CtoPData(response)
	if err != nil {
		mh.log.Errorf("Error converting conversation entity to message payload for request id: %s with error : %v", incomingMsg.Id, err)
		return &messages.MessageResponse{
			Workflow: incomingMsg.Workflow,
			Status:   503,
			ErrorMsg: err.Error(),
		}
	}
	return &messages.MessageResponse{
		Workflow: incomingMsg.Workflow,
		Status:   200,
		Response: dataMap,
	}
}

func (mh *MsgHandler) PDataToC(data map[string]interface{}) (*dhauli.Conversation, error) {
	var conversation dhauli.Conversation
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		mh.log.Errorf("Error marshalling Payload data: %v to JSON", err)
		return nil, err
	}
	if err = json.Unmarshal(jsonBytes, &conversation); err != nil {
		mh.log.Errorf("Error unmarshalling Payload JSON data: %v to Conversation struct", err)
		return nil, err
	}
	return &conversation, nil
}

func (mh *MsgHandler) CtoPData(conv *dhauli.Conversation) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(conv)
	if err != nil {
		mh.log.Errorf("Error marshalling Conversation struct: %v to JSON", err)
		return nil, err
	}
	var dataMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &dataMap)
	if err != nil {
		mh.log.Errorf("Error unmarshalling JSON to map: %v", err)
		return nil, err
	}
	return dataMap, nil
}
