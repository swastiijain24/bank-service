package workers

import (
	"context"
	"fmt"
	"log"

	Kafka "github.com/segmentio/kafka-go"
	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"github.com/swastiijain24/bank/internals/services"
	"google.golang.org/protobuf/proto"
)

const maxRetryCount = 5

type BankWorker struct {
	bankConsumer  *kafka.Consumer
	dlqProducer   *kafka.Producer
	bankService   services.BankService
	failureCounts map[string]int
}

func NewBankWorker(bankConsumer *kafka.Consumer, dlqProducer *kafka.Producer, bankService services.BankService) *BankWorker {
	return &BankWorker{
		bankConsumer:  bankConsumer,
		dlqProducer:   dlqProducer,
		bankService:   bankService,
		failureCounts: make(map[string]int),
	}
}

func (w *BankWorker) Start(ctx context.Context) {

	for {

		msg, err := w.bankConsumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Fetch error: %v", err)
			continue
		}

		var bankPayment pb.BankRequest

		err = proto.Unmarshal(msg.Value, &bankPayment)
		if err != nil {
			log.Printf("error unpacking message: %v", err)
			if err := w.moveToDLQ(ctx, msg, err.Error()); err == nil {
				w.clearFailureCount(msg)
			}
			continue
		}

		log.Print("request getting processed by the bank 7")

		err = w.bankService.ExecuteBankOperation(ctx, &bankPayment)

		if err != nil {
			log.Printf("failed to execute bank operation: %v", err)
			attempt := w.incrementFailureCount(msg)
			if attempt >= maxRetryCount {
				reason := fmt.Sprintf("transient retries exceeded (%d): %v", attempt, err)
				if err := w.moveToDLQ(ctx, msg, reason); err == nil {
					w.clearFailureCount(msg)
				}
				continue
			}

			log.Printf("transient processing error for key=%s attempt=%d/%d: %v", string(msg.Key), attempt, maxRetryCount, err)
			continue
		}

		log.Print("executed bank operation")

		if err := w.bankConsumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit: %v", err)
			continue
		}
		w.clearFailureCount(msg)

	}

}

func (w *BankWorker) moveToDLQ(ctx context.Context, msg Kafka.Message, reason string) error {
	log.Printf("Moving message %s to DLQ. Reason: %s", string(msg.Key), reason)
	err := w.dlqProducer.ProduceEvent(ctx, string(msg.Key), msg.Value, "bank.response.failed")
	if err != nil {
		log.Printf("DLQ write failed for key=%s: %v", string(msg.Key), err)
		return err
	}

	if err := w.bankConsumer.Reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("Failed to commit poisoned message: %v", err)
		return err
	}

	return nil
}

func (w *BankWorker) messageID(msg Kafka.Message) string {
	return fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)
}

func (w *BankWorker) incrementFailureCount(msg Kafka.Message) int {
	id := w.messageID(msg)
	w.failureCounts[id]++
	return w.failureCounts[id]
}

func (w *BankWorker) clearFailureCount(msg Kafka.Message) {
	delete(w.failureCounts, w.messageID(msg))
}
