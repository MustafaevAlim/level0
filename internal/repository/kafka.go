package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator"
	"github.com/segmentio/kafka-go"

	"level0/internal/model"
)

func RegisterDateTimeValidation(validate *validator.Validate) {
	err := validate.RegisterValidation("notzero", func(fl validator.FieldLevel) bool {
		t, ok := fl.Field().Interface().(time.Time)
		if !ok {
			return false
		}
		return !t.IsZero()
	})
	if err != nil {
		log.Println("Ошибка регистрации валидации времени")
	}

}

type KafkaReader struct {
	Reader *kafka.Reader
}

func NewKafkaReader(brokers []string, topic string, groupId string) *KafkaReader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupId,
		CommitInterval: 0,
	})
	return &KafkaReader{Reader: reader}
}

func (r *KafkaReader) Consume(ctx context.Context, ch chan model.OrderMsg) error {

	defer close(ch)

	validate := validator.New()
	RegisterDateTimeValidation(validate)

	for {
		msg, err := r.Reader.ReadMessage(ctx)
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

		err = validate.Struct(data)
		if err != nil {
			log.Printf("Кафка: ошибка валидации данных: %v", err)
			continue
		}

		log.Printf("Кафка: получен заказ с айди %s", data.OrderUID)

		select {
		case ch <- model.OrderMsg{Msg: msg, Order: data}:
		case <-ctx.Done():
			return ctx.Err()
		}

	}

}
