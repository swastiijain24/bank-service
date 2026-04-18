package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	httpclient "github.com/swastiijain24/bank/internals/http_client"
	"github.com/swastiijain24/bank/internals/kafka"
	"github.com/swastiijain24/bank/internals/repository"
	"github.com/swastiijain24/bank/internals/services"
	"github.com/swastiijain24/bank/internals/workers"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := godotenv.Load()
	if err != nil {
		log.Print("no .env file found")
	}

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	responseProducer := kafka.NewProducer(kafkaAddr)

	bankClient := httpclient.NewBankClient(os.Getenv("REQ_URL"))

	redis := repository.NewRedisStore(os.Getenv("REDIS_ADDR"),  24*time.Hour)
	bankSvc := services.NewBankService(responseProducer, redis, bankClient)

	consumer := kafka.NewConsumer([]string{kafkaAddr}, "bank.instruction.v1", "bank-grp")
	defer consumer.Reader.Close()

	dlqProducer := kafka.NewProducer(kafkaAddr)
	bankWorker := workers.NewBankWorker(consumer, dlqProducer, bankSvc)
	statusConsumer := kafka.NewConsumer([]string {kafkaAddr}, "bank.enquiry.v1", "bank-grp2")
	statusWorker := workers.NewStatusWorker(statusConsumer, bankSvc)

	log.Println("Bank Service Worker started...")
	go bankWorker.Start(ctx)
	go statusWorker.StartStatusWorker(ctx)
	<-ctx.Done()

	log.Println("Shutting down...")

}
