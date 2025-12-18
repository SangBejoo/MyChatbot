package interfaces

import "project_masAde/internal/entities"

type AIClient interface {
	GenerateResponse(prompt string) (string, error)
}

type Messenger interface {
	SendMessage(to, content string) error
	ReceiveMessage() (entities.Message, error)
}