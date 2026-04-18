package workers

import (
	"context"
	"log"

	"github.com/swastiijain24/bank/internals/kafka"
	"github.com/swastiijain24/bank/internals/services"
)

type StatusWorker struct {
	consumer *kafka.Consumer
	bankService services.BankService
}

func NewStatusWorker(consumer *kafka.Consumer, bankService services.BankService) *StatusWorker {
	return &StatusWorker{
		consumer: consumer,
		bankService: bankService,
	}
}

func (w *StatusWorker) StartStatusWorker(ctx context.Context) {
	for {
		msg, err := w.consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Fetch error: %v", err)
			continue
		}

		err = w.bankService.CheckStatus(ctx, string(msg.Key), string(msg.Value))
		if err != nil {
			log.Printf("failed to fetch status : %v", err)
			continue 
		}

		if err := w.consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit: %v", err)
		}

	}
}