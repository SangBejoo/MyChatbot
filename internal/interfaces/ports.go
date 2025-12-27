package interfaces

import "project_masAde/internal/entities"

// AIClient removed - AI handled by separate microservice
// See: AI Service project for AI integration

type Messenger interface {
	SendMessage(to, content string) error
	ReceiveMessage() (entities.Message, error)
}