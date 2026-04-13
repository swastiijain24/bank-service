package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"google.golang.org/protobuf/proto"
)

type BankService interface {
	ExecuteBankOperation(ctx context.Context, transactionId string, payerAccountId string, payeeAccountId string, amount int64, Type string, bankCode string) error 
}

type banksvc struct {
	Producer *kafka.Producer
}

func NewBankService(producer *kafka.Producer) BankService {
	return &banksvc{
		Producer: producer,
	}
}

func (s *banksvc) ExecuteBankOperation(ctx context.Context, transactionId string, payerAccountId string, payeeAccountId string, amount int64, Type string, bankCode string) error {

	//the actual txn response returned by bank is this in json
	//we will use this txn id returned by the bank as the bank refernce id which is there in the bankreponse proto
	//also for sending req to banks apis we will have to  set idempotency and api key for idem[otency key we can use the txn id and api key will be genrated by bank and we will use it
	// 	type Transaction struct {
	// 	ID                  pgtype.UUID        `json:"id"`
	// 	FromAccountID       string             `json:"from_account_id"`
	// 	ToAccountIdentifier string             `json:"to_account_identifier"`
	// 	Amount              int64              `json:"amount"`
	// 	Status              string             `json:"status"`
	// 	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	// 	UpdatedAt           pgtype.Timestamptz `json:"updated_at"`
	// }

	//if type is credit we will call the credit api of that bank else debit or if anything else like seeing the balance or creating an acc then unknown

	//based on the bank code we will identify which bank to call 

	log.Print("received request")
	log.Print("11")
	var bankResponse *pb.BankResponse

	switch Type {

	case "DEBIT":
		//call debit api
		bankTransactionResponseId := bankresponse()
		message := &pb.BankResponse{
			TransactionId:   transactionId,
			BankReferenceId: bankTransactionResponseId,
			Success:         true,
			ErrorMessage:    "",
			Type:            "DEBIT",
		}
		bankResponse = message

	case "CREDIT":
		//call credit api
		bankTransactionResponseId := bankresponse()
		message := &pb.BankResponse{
			TransactionId:   transactionId,
			BankReferenceId: bankTransactionResponseId,
			Success:         true,
			ErrorMessage:    "",
			Type:            "CREDIT",
		}
		bankResponse = message

	default:
		//call another getcreateaccountapi
		bankTransactionResponseId := bankresponse()
		message := &pb.BankResponse{
			TransactionId:   transactionId,
			BankReferenceId: bankTransactionResponseId,
			Success:         true,
			ErrorMessage:    "",
			Type:            "UNKNOWN",
		}
		bankResponse = message
	}

	data, err := proto.Marshal(bankResponse)
	if err != nil {
		return fmt.Errorf("error packing data : %w", err)
	}

	s.Producer.ProduceEvent(ctx, transactionId, data)
	

	log.Print("12")
	return nil 
}

func bankresponse() string {
	return "hjfvhfv"+strconv.Itoa(rand.Int())
}