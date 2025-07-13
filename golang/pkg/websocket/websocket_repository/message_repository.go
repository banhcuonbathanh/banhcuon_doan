package websocket_repository

import (
	websocket_model "english-ai-full/pkg/websocket/websocker_model"
)

type MessageRepository interface {
	SaveMessage(message *websocket_model.Message) error
	GetMessages() ([]*websocket_model.Message, error)
}

type inMemoryMessageRepository struct {
	messages []*websocket_model.Message
}

func NewInMemoryMessageRepository() MessageRepository {
	return &inMemoryMessageRepository{
		messages: make([]*websocket_model.Message, 0),
	}
}

func (r *inMemoryMessageRepository) SaveMessage(message *websocket_model.Message) error {
	r.messages = append(r.messages, message)
	return nil
}

func (r *inMemoryMessageRepository) GetMessages() ([]*websocket_model.Message, error) {
	return r.messages, nil
}
