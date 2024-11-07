package messagepublisher

import (
	"context"
	"sync"

	"cosmic-go/internal/domain"
	unitofwork "cosmic-go/internal/uow"

	"github.com/redis/go-redis/v9"
)

type MessagePublisher struct {
	eventHandlers  map[string][]unitofwork.EventHandler
	commandHandler map[string]unitofwork.CommandHandler
}

func NewMessagePublisher() *MessagePublisher {
	return &MessagePublisher{}
}

func (e *MessagePublisher) RegisterEventHandler(
	event domain.Event,
	handler unitofwork.EventHandler,
) {
	if e.eventHandlers == nil {
		e.eventHandlers = make(map[string][]unitofwork.EventHandler)
	}

	e.eventHandlers[event.GetEventName()] = append(e.eventHandlers[event.GetEventName()], handler)
}

func (mp *MessagePublisher) PublishEvent(event domain.Event) <-chan error {
	if eventHandlers, ok := mp.eventHandlers[event.GetEventName()]; ok {
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

func (mp *MessagePublisher) RegisterCommandHandler(
	command domain.Command,
	handler unitofwork.CommandHandler,
) {
	if mp.commandHandler == nil {
		mp.commandHandler = make(map[string]unitofwork.CommandHandler)
	}

	mp.commandHandler[command.GetCommandName()] = handler
}

func (mp *MessagePublisher) PublishCommand(command domain.Command) <-chan error {
	if handler, ok := mp.commandHandler[command.GetCommandName()]; ok {
		errChan := make(chan error, 1)

		if err := handler(command); err != nil {
			errChan <- err
		}

		close(errChan)
		return errChan
	}
	return nil
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
	if err := emp.client.Publish(
		context.Background(),
		event.GetEventName(),
		event,
	).Err(); err != nil {
		return err
	}

	return nil
}
