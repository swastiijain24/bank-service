package workers

import (
	"context"
	"log"

	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"github.com/swastiijain24/bank/internals/services"
	"google.golang.org/protobuf/proto"
)

type BankWorker struct {
	bankConsumer *kafka.Consumer
	bankService  services.BankService
}

func NewBankWorker(bankConsumer *kafka.Consumer, bankService services.BankService) *BankWorker {
	return &BankWorker{
		bankConsumer: bankConsumer,
		bankService:  bankService,
	}
}

func (w *BankWorker) Start(ctx context.Context) {

	for {

		msg, err := w.bankConsumer.Reader.FetchMessage(ctx)
		if err != nil {
			break
		}

		var bankPayment pb.BankRequest

		err = proto.Unmarshal(msg.Value, &bankPayment)
		if err != nil {
			log.Printf("error unpacking message: %v", err)
			continue
		}
		log.Print("10")
		err = w.bankService.ExecuteBankOperation(ctx, bankPayment.GetTransactionId(), bankPayment.GetPayerAccountId(), bankPayment.GetPayeeAccountId(), bankPayment.GetAmount(), bankPayment.GetType(), bankPayment.GetBankCode())
		if err != nil {
			log.Printf("failed to execute bank operation: %v", err)
		}
		log.Print("13")
		if err := w.bankConsumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit: %v", err)
		}
	}

}
