package workers

import (
	"context"
	"fmt"

	pb "github.com/swastiijain24/bank/internals/gen"
	"github.com/swastiijain24/bank/internals/kafka"
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

		msg, err := w.bankConsumer.Reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var bankPayment pb.BankRequest

		err = proto.Unmarshal(msg.Value, &bankPayment)
		if err != nil {
			fmt.Println("error unpacking message:", err)
			continue
		}

		w.bankService.ExecuteBankOperation(ctx, bankPayment.GetTransactionId(), bankPayment.GetPayerAccountId(), bankPayment.GetPayeeAccountId(), bankPayment.GetAmount(), bankPayment.GetType())

	}

}
