package handler

import (
	"context"
	"log"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/messagepublisher"
	"cosmic-go/pkg/utils"

	"github.com/redis/go-redis/v9"
)

type ExternalMessageConsumer struct {
	client *redis.Client
	imp    *messagepublisher.MessagePublisher
}

func NewExternalMessageConsumer(
	client *redis.Client,
	imp *messagepublisher.MessagePublisher,
) *ExternalMessageConsumer {
	return &ExternalMessageConsumer{
		client: client,
		imp:    imp,
	}
}

func (emc *ExternalMessageConsumer) ConsumeChangeBatchQuantityCommand(ctx context.Context) {
	command := domain.ChangeBatchQuantity{}
	ch := emc.client.Subscribe(ctx, command.GetCommandName()).Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			command, err := domain.NewChangeBatchQuantityFromJson(msg.Payload)
			if err != nil {
				log.Printf("failed to read command from external pub/sub: %v", err)
				return
			}

			errChan := emc.imp.PublishCommand(command)
			utils.LogErrChan(errChan)
		}
	}
}
