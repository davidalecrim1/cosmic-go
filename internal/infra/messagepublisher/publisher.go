package messagepublisher

import (
	"context"
	"encoding/json"
	"sync"

	"cosmic-go/internal/domain"
	unitofwork "cosmic-go/internal/uow"

	"github.com/redis/go-redis/v9"
)

type InternalMessagePublisher struct {
	eventHandlers  map[string][]unitofwork.EventHandler
	commandHandler map[string]unitofwork.CommandHandler
}

func NewInternalMessagePublisher() *InternalMessagePublisher {
	return &InternalMessagePublisher{}
}

func (imp *InternalMessagePublisher) RegisterEventHandler(
	event domain.Event,
	handler unitofwork.EventHandler,
) {
	if imp.eventHandlers == nil {
		imp.eventHandlers = make(map[string][]unitofwork.EventHandler)
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
	handler unitofwork.CommandHandler,
) {
	if imp.commandHandler == nil {
		imp.commandHandler = make(map[string]unitofwork.CommandHandler)
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

type ExternalMessagePublisher struct {
	client *redis.Client
}

func NewExternalMessagePublisher(client *redis.Client) *ExternalMessagePublisher {
	return &ExternalMessagePublisher{
		client: client,
	}
}

func (emp *ExternalMessagePublisher) PublishEvent(event domain.Event) error {
	eventAsJson, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := emp.client.Publish(
		context.Background(),
		event.GetEventName(),
		eventAsJson,
	).Err(); err != nil {
		return err
	}

	return nil
}
