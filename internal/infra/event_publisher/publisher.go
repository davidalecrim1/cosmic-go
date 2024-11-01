package eventpublisher

import (
	"sync"

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

func (ep *EventPublisher) Publish(event domain.Event) <-chan error {
	if handlers, ok := ep.handlers[event.EventName()]; ok {
		errChan := make(chan error, len(handlers))

		var wg sync.WaitGroup
		for _, handler := range handlers {
			wg.Add(1)
			go func(e domain.Event) {
				defer wg.Done()
				if err := handler(event); err != nil {
					errChan <- err
				}
			}(event)
		}

		wg.Wait()
		close(errChan)
		return errChan
	}
	return nil
}
