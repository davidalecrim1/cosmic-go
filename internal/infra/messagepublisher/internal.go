package messagepublisher

import (
	"sync"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
)

type InternalMessagePublisher struct {
	eventHandlers  map[string][]application.EventHandler
	commandHandler map[string]application.CommandHandler
}

func NewInternalMessagePublisher() *InternalMessagePublisher {
	return &InternalMessagePublisher{}
}

func (imp *InternalMessagePublisher) RegisterEventHandler(
	event domain.Event,
	handler application.EventHandler,
) {
	if imp.eventHandlers == nil {
		imp.eventHandlers = make(map[string][]application.EventHandler)
	}

	imp.eventHandlers[event.GetEventName()] = append(imp.eventHandlers[event.GetEventName()], handler)
}

func (imp *InternalMessagePublisher) PublishEvent(event domain.Event) <-chan error {
	if eventHandlers, ok := imp.eventHandlers[event.GetEventName()]; ok {
		errChan := make(chan error, len(eventHandlers))

		var wg sync.WaitGroup
		for _, handler := range eventHandlers {
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

func (imp *InternalMessagePublisher) RegisterCommandHandler(
	command domain.Command,
	handler application.CommandHandler,
) {
	if imp.commandHandler == nil {
		imp.commandHandler = make(map[string]application.CommandHandler)
	}

	imp.commandHandler[command.GetCommandName()] = handler
}

func (imp *InternalMessagePublisher) PublishCommand(command domain.Command) <-chan error {
	errChan := make(chan error, 1)

	if handler, ok := imp.commandHandler[command.GetCommandName()]; ok {
		if err := handler(command); err != nil {
			errChan <- err
		}
	}

	close(errChan)
	return errChan
}
