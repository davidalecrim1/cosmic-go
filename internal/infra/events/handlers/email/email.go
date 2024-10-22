package email

import (
	"log"

	"cosmic-go/internal/domain"
)

type EmailService struct{}

func (e *EmailService) Handle(event domain.Event) error {
	log.Printf("simulating an e-mail being sent for event: %v", event.EventName())
	return nil
}
