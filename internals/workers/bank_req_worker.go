package workers

import (
	"context"
	Kafka "github.com/segmentio/kafka-go"
	"log"
	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"github.com/swastiijain24/bank/internals/services"
	"google.golang.org/protobuf/proto"
)

type BankWorker struct {
	bankConsumer *kafka.Consumer
	dlqProducer *kafka.Producer
	bankService  services.BankService
}

func NewBankWorker(bankConsumer *kafka.Consumer,dlqProducer *kafka.Producer, bankService services.BankService) *BankWorker {
	return &BankWorker{
		bankConsumer: bankConsumer,
		dlqProducer: dlqProducer,
		bankService:  bankService,
	}
}

func (w *BankWorker) Start(ctx context.Context) {

	for {

		msg, err := w.bankConsumer.Reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("error fetching message: %v", err)
			continue
		}

		var bankPayment pb.BankRequest

		err = proto.Unmarshal(msg.Value, &bankPayment)
		if err != nil {
			log.Printf("error unpacking message: %v", err)
			w.moveToDLQ(ctx,msg, err.Error() )
			continue
		}

		log.Print("request getting processed by the bank")

		err = w.bankService.ExecuteBankOperation(ctx, &bankPayment)

		if err != nil {
			log.Printf("failed to execute bank operation: %v", err)
			continue 
		}

		log.Print("executed bank operation")

		if err := w.bankConsumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit: %v", err)
		}

	}

}

func (w *BankWorker) moveToDLQ(ctx context.Context, msg Kafka.Message, reason string) {
	log.Printf("Moving message %s to DLQ. Reason: %s", string(msg.Key), reason)
	err := w.dlqProducer.ProduceEvent(ctx, string(msg.Key), msg.Value, "bank.response.failed")
	if err != nil {
		log.Fatalf("Critical Failure: Cannot write to DLQ: %v", err)
	}
	

	if err := w.bankConsumer.Reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("Failed to commit poisoned message: %v", err)
	}
}

