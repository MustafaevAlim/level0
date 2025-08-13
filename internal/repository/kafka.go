package repository

import (
	"Level0/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	reader *kafka.Reader
}

func NewKafkaReader(brokers []string, topic string, groupId string) *KafkaReader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupId,
	})
	return &KafkaReader{reader: reader}
}

func (r *KafkaReader) Consume(ctx context.Context, ch chan model.Order) error {

	defer r.reader.Close()
	defer close(ch)

	for {
		msg, err := r.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Кафка: отмена контекста")
				return nil
			}
			return fmt.Errorf("ошибка получения сообщения: %w", err)
		}
		var data model.Order
		err = json.Unmarshal(msg.Value, &data)

		if err != nil {
			log.Printf("Кафка: ошибка десериализации: %v", err)
			continue
		}
		log.Printf("Кафка: получен заказ с айди %s", data.OrderUID)

		select {
		case ch <- data:
		case <-ctx.Done():
			return ctx.Err()
		}

	}

}
