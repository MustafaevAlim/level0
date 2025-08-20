package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func generateStringInt(n int) string {
	digits := "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

func generateRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomPhone() string {
	return fmt.Sprintf("+%d%09d", rand.Intn(90)+10, rand.Intn(1_000_000_000))
}

func generateRandomOrder() Order {
	now := time.Now()
	return Order{
		OrderUID:          generateRandomString(16),
		TrackNumber:       "TRK" + generateRandomString(8),
		Entry:             generateRandomString(4),
		Locale:            "en",
		InternalSignature: generateRandomString(3),
		CustomerID:        generateRandomString(6),
		DeliveryService:   generateRandomString(5),
		ShardKey:          generateStringInt(2),
		SmID:              rand.Intn(1000),
		DateCreated:       now,
		OofShard:          generateStringInt(2),
		Delivery: Delivery{
			Name:    generateRandomString(6),
			Phone:   randomPhone(),
			Zip:     generateStringInt(5),
			City:    generateRandomString(14),
			Address: generateRandomString(10),
			Region:  generateRandomString(6),
			Email:   generateRandomString(7) + "@gmail.com",
		},
		Payment: Payment{
			Transaction:  generateRandomString(16),
			RequestID:    generateRandomString(5),
			Currency:     "USD",
			Provider:     generateRandomString(12),
			Amount:       rand.Intn(10000),
			PaymentDT:    rand.Int63(),
			Bank:         generateRandomString(6),
			DeliveryCost: rand.Intn(1_000_000),
			GoodsTotal:   rand.Intn(1_000_000),
			CustomFee:    rand.Intn(1_000_000),
		},
		Items: []Item{{
			ChrtID:      rand.Intn(1_000_000),
			TrackNumber: "TRK" + generateRandomString(8),
			Price:       rand.Intn(1_000_000),
			Rid:         generateRandomString(16),
			Name:        generateRandomString(6),
			Sale:        rand.Intn(101),
			Size:        generateRandomString(2),
			TotalPrice:  rand.Intn(1_000_000),
			NmID:        rand.Intn(1_000_000),
			Brand:       generateRandomString(13),
			Status:      202,
		}},
	}
}

func WriteInKafka() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "orders-topic",
	})
	defer func() {
		if err := writer.Close(); err != nil {
			log.Println("Ошибка при закрытии соединения продюсера с Кафкой: ", err)
		}
	}()

	for {
		order := generateRandomOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Fatalln("Ошибка в генерации тестовых данных: ", err)
		}

		err = writer.WriteMessages(context.Background(), kafka.Message{
			Value: data,
		})
		if err != nil {
			log.Println("Kafka недоступна, повтор через 5 секунд:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		time.Sleep(time.Duration(rand.Intn(60)) * time.Second)

	}
}
