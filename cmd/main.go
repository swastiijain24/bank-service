package main

import (
	"github.com/swastiijain24/bank/internals/kafka"
	"github.com/swastiijain24/bank/internals/services"
)

func main() {

	
	bankOutcomeProducer := kafka.NewProducer("localhost:9092", "bank.outcome.v1")
	bankService := services.NewBankService(bankOutcomeProducer)
	
}
