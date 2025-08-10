package main

import (
	"Level0/internal/app"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

// Загрузка переменных окружения
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	data, err := os.ReadFile("model.json")
	if err != nil {
		log.Fatal(err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders-topic",
	})
	defer writer.Close()

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: data,
	})
	if err != nil {
		log.Fatalln(err)
	}

	a := app.NewApp()
	a.Run()

}
