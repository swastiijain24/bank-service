package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"github.com/swastiijain24/bank/internals/repository"
	"google.golang.org/protobuf/proto"
)

type BankService interface {
	ExecuteBankOperation(ctx context.Context, bankRequest *pb.BankRequest) error
	CheckStatus(ctx context.Context, transactionId string) error 
}

type banksvc struct {
	Producer *kafka.Producer
	redis    *repository.RedisStore
}

func NewBankService(producer *kafka.Producer, redis *repository.RedisStore) BankService {
	return &banksvc{
		Producer: producer,
		redis:    redis,
	}
}

func (s *banksvc) ExecuteBankOperation(ctx context.Context, bankRequest *pb.BankRequest) error {

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

	// will have to use a timeout context to call the apis because the system will hang if it  takes too long for the bank to process
	var opType string
	switch bankRequest.Operation.(type) {
	case *pb.BankRequest_Debit:
		opType = "DEBIT"
	case *pb.BankRequest_Credit:
		opType = "CREDIT"
	default:
		opType = "UNKNOWN"
	}

	redisKey := "idempotency:" + bankRequest.GetTransactionId() + opType
	cachedData, err := s.redis.GetRedisResponse(redisKey)
	if err != nil {
		log.Printf("redis error: %v", err)
	}
	if cachedData != nil {
		return s.Producer.ProduceEvent(ctx, bankRequest.GetTransactionId(), cachedData, "bank.response.v1")
	}

	bankTransactionResponseId := bankcall()
	bankResponse := &pb.BankResponse{
		TransactionId:   bankRequest.GetTransactionId(),
		BankReferenceId: bankTransactionResponseId,
		Success:         true,
		ErrorMessage:    "",
		Type:            opType,
	}

	data, err := proto.Marshal(bankResponse)
	if err != nil {
		return fmt.Errorf("error packing data : %w", err)
	}

	if bankResponse.Success {
		err = s.redis.SetRedisResponse(redisKey, data)
		if err != nil {
			log.Printf("failed to save to redis %v", err)
		}
	}

	err = s.Producer.ProduceEvent(ctx, bankRequest.GetTransactionId(), data, "bank.response.v1")
	if err != nil {
		return err
	}

	log.Print("produced bank response")
	return nil

}

func bankcall() string {

	return "hjfvhfv" + strconv.Itoa(rand.Int())
}

func (s *banksvc) CheckStatus(ctx context.Context, transactionId string) error {

	//call the get status api of the bank where the external id in the core bank model will be this transaction id
	//based on the response we will set the status
	//it will return the status and the bank ref id if pending or fail then success = false else true
	bankTransactionResponseId := bankcall()
	var opType string //this will help us know either debit pending or credit pending so basically we will be sending debit suc/fail or credit suc/fail
	bankResponse := &pb.BankResponse{
		TransactionId:   transactionId,
		BankReferenceId: bankTransactionResponseId,
		Success:         false,
		ErrorMessage:    "",
		Type:            opType,
	}

	data, err := proto.Marshal(bankResponse)
	if err != nil {
		return fmt.Errorf("error packing data : %w", err)
	}

	err = s.Producer.ProduceEvent(ctx, transactionId, data, "bank.response.v1")
	if err != nil {
		return fmt.Errorf("error producing event : %w", err)
	}
	log.Println("transaction status produced")
	return nil 
}
