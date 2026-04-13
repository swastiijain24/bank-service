package main

import (
	"context"
	"log"

	"github.com/swastiijain24/bank/internals/kafka"
	"github.com/swastiijain24/bank/internals/services"
	"github.com/swastiijain24/bank/internals/workers"
)

func main() {
	respProducer := kafka.NewProducer("localhost:9092", "bank.response.v1")
	
	bankSvc := services.NewBankService(respProducer)

	consumer := kafka.NewConsumer([]string {"localhost:9092"}, "bank.instruction.v1")
	bankWorker := workers.NewBankWorker(consumer, bankSvc)

	log.Println("Bank Service Worker started...")
	bankWorker.Start(context.Background())
	// select {}
}