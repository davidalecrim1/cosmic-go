package messagepublisher

import (
	"context"
	"encoding/json"

	"cosmic-go/internal/domain"

	"github.com/redis/go-redis/v9"
)

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
