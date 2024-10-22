package publisher

import (
	"log"

	"cosmic-go/internal/domain"
	unitofwork "cosmic-go/internal/uow"
)

type EventPublisher struct {
	handlers map[string][]unitofwork.EventHandler
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (e *EventPublisher) RegisterHandler(
	event domain.Event,
	handler unitofwork.EventHandler,
) {
	if e.handlers == nil {
		e.handlers = make(map[string][]unitofwork.EventHandler)
	}

	e.handlers[event.EventName()] = append(e.handlers[event.EventName()], handler)
}

func (e *EventPublisher) Publish(event domain.Event) {
	if handlers, ok := e.handlers[event.EventName()]; ok {
		for _, handler := range handlers {
			err := handler.Handle(event)
			if err != nil {
				log.Printf("failed to process event %v with error: %v", event, err)
			}
		}
	}
}
