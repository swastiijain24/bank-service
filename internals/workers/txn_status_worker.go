package workers

import (
	"context"
	"fmt"
	"log"

	"github.com/swastiijain24/bank/internals/kafka"
	"github.com/swastiijain24/bank/internals/services"
)

type StatusWorker struct {
	consumer *kafka.Consumer
	bankService services.BankService
}

func NewStatusWorker(consumer *kafka.Consumer) *StatusWorker {
	return &StatusWorker{
		consumer: consumer,
	}
}

func (w *StatusWorker) StartStatusWorker(ctx context.Context) {
	for {
		msg, err := w.consumer.Reader.FetchMessage(ctx)
		if err != nil {
			fmt.Println("error fetching message:", err)
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