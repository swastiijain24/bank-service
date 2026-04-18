package services

import (
	"context"
	"errors"
	"fmt"
	httpclient "github.com/swastiijain24/bank/internals/http_client"
	"github.com/swastiijain24/bank/internals/kafka"
	pb "github.com/swastiijain24/bank/internals/pb"
	"github.com/swastiijain24/bank/internals/repository"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"os"
)

type BankService interface {
	ExecuteBankOperation(ctx context.Context, bankRequest *pb.BankRequest) error
	CheckStatus(ctx context.Context, transactionId string, transactionType string) error
}

type banksvc struct {
	Producer   *kafka.Producer
	redis      *repository.RedisStore
	bankClient *httpclient.BankClient
}

func NewBankService(producer *kafka.Producer, redis *repository.RedisStore, bankClient *httpclient.BankClient) BankService {
	return &banksvc{
		Producer:   producer,
		redis:      redis,
		bankClient: bankClient,
	}
}

func (s *banksvc) CreateBankResponse(transactionId string, bankRefId string, status string, errMsg string, opType string) *pb.BankResponse {

	return &pb.BankResponse{
		TransactionId:   transactionId,
		BankReferenceId: bankRefId,
		Success:         (status == "SUCCESS"),
		ErrorMessage:    errMsg,
		Type:            opType,
	}
}

func (s *banksvc) CheckCache(ctx context.Context, redisKey string, transactionId string) (bool, error) {
	cachedData, err := s.redis.GetRedisResponse(redisKey)
	if err != nil {
		log.Printf("redis error: %v", err)
		return false, err
	}
	if cachedData != nil {
		return true, s.Producer.ProduceEvent(ctx, transactionId, cachedData, "bank.response.v1")
	}
	return false, nil
}

func (s *banksvc) ExecuteBankOperation(ctx context.Context, bankRequest *pb.BankRequest) error {

	var opType string
	var redisKey string
	errMsg := ""
	var bankResponse *pb.BankResponse

	switch bankRequest.Operation.(type) {
	case *pb.BankRequest_Debit:
		
		opType = "DEBIT"
		redisKey = "idempotency:" + bankRequest.GetTransactionId() + opType
		cached, err := s.CheckCache(ctx, redisKey, bankRequest.GetTransactionId())
		if cached == true {
			return nil
		}
		if err != nil {
			return fmt.Errorf("cache error: %w", err)
		}

		log.Print("calling debit api 8")

		bankRefernceId, status, err := s.bankClient.CallDebit(ctx, bankRequest.TransactionId, bankRequest.PayerAccountId, bankRequest.PayeeAccountId, bankRequest.Amount, bankRequest.GetDebit().Mpin)
		if err != nil {
			if isTemporary(err) {
				log.Printf("network timeout for txn %s", err)
				return nil
			}
			errMsg = err.Error()
		}
		log.Print("debit from main bank done 13")
		bankResponse = s.CreateBankResponse(bankRequest.GetTransactionId(), bankRefernceId, status, errMsg, "DEBIT")

	case *pb.BankRequest_Credit:

		opType = "CREDIT"
		redisKey = "idempotency:" + bankRequest.GetTransactionId() + opType
		cached, err := s.CheckCache(ctx, redisKey, bankRequest.GetTransactionId())
		if cached == true {
			return nil
		}
		if err != nil {
			return fmt.Errorf("cache error: %w", err)
		}
		log.Print("calling credit api 17")
		bankRefernceId, status, err := s.bankClient.CallCredit(ctx, bankRequest.TransactionId, bankRequest.PayerAccountId, bankRequest.PayeeAccountId, bankRequest.Amount)
		if err != nil {
			if isTemporary(err) {
				log.Printf("network timeout for txn %s", err)
				return nil
			}
			errMsg = err.Error()
		}
		log.Print("credit done 23")
		bankResponse = s.CreateBankResponse(bankRequest.GetTransactionId(), bankRefernceId, status, errMsg, "CREDIT")

	case *pb.BankRequest_Refund:
		opType = "REFUND"
		redisKey = "idempotency:" + bankRequest.GetTransactionId() + opType
		cached, err := s.CheckCache(ctx, redisKey, bankRequest.GetTransactionId())
		if cached == true {
			return nil
		}
		if err != nil {
			return fmt.Errorf("cache error: %w", err)
		}
		bankRefernceId, status, err := s.bankClient.CallRefund(ctx, bankRequest.TransactionId, bankRequest.PayerAccountId, bankRequest.PayeeAccountId, bankRequest.Amount)
		if err != nil {
			if isTemporary(err) {
				log.Printf("network timeout for txn %s", err)
				return nil
			}
			errMsg = err.Error()
		}
		bankResponse = s.CreateBankResponse(bankRequest.GetTransactionId(), bankRefernceId, status, errMsg, "REFUND")
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

	log.Print("produced bank response 14")
	log.Print("produced response 24")
	return nil

}

func (s *banksvc) CheckStatus(ctx context.Context, transactionId string, transactionType string) error {

	//call the get status api of the bank where the external id in the core bank model will be this transaction id, based on the response we will set the status, it will return the status and the bank ref id if pending or fail then success = false else true
	bankReferenceId, status, err := s.bankClient.GetStatusFromBank(ctx, transactionId, transactionType)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	bankResponse := &pb.BankResponse{
		TransactionId:   transactionId,
		BankReferenceId: bankReferenceId,
		Success:         (status == "SUCCESS"),
		ErrorMessage:    errMsg,
		Type:            transactionType,
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

func isTemporary(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if os.IsTimeout(err) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}
