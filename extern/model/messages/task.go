package messages

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
)

type MessageTaskResult struct {
	Messages          Messages
	InternalMessages  InternalMessages
	MessageGasEconomy *MessageGasEconomy
}

func (mtr *MessageTaskResult) Persist(ctx context.Context, s model.StorageBatch) error {
	if err := mtr.Messages.Persist(ctx, s); err != nil {
		return err
	}
	if err := mtr.InternalMessages.Persist(ctx, s); err != nil {
		return err
	}
	if err := mtr.MessageGasEconomy.Persist(ctx, s); err != nil {
		return err
	}
	return nil
}
