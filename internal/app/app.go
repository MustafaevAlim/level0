package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"level0/internal/config"
	"level0/internal/model"
	"level0/internal/repository"
)

type App struct {
	DB     *repository.Storage
	Cache  *repository.LRUcache
	Kafka  *repository.KafkaReader
	Router http.Handler
}

func NewApp(db *repository.Storage, cache *repository.LRUcache, kafka *repository.KafkaReader, router http.Handler) *App {

	return &App{DB: db, Cache: cache, Kafka: kafka, Router: router}
}

func (a *App) Run(ctx context.Context) error {
	conf := config.New()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	orders := make(chan model.OrderMsg, 100)
	go func() {
		defer wg.Done()
		if err := a.Kafka.Consume(ctx, orders); err != nil {
			log.Printf("Kafka error: %v", err)
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		for orderMsg := range orders {
			if err := a.DB.AddOrder(ctx, orderMsg.Order); err != nil {
				log.Printf("DB error: %v", err)
				continue
			}

			a.Cache.Push(orderMsg.Order)

			if err := a.Kafka.Reader.CommitMessages(ctx, orderMsg.Msg); err != nil {
				log.Printf("Не удалось закоммитить offset: %v", err)

			}

		}
	}()

	server := &http.Server{
		Handler: a.Router,
		Addr:    ":" + conf.Server.Port,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP error: %v", err)
			cancel()
		}
	}()
	<-ctx.Done()

	log.Println("Завершение программы...")
	wg.Wait()

	err := server.Shutdown(context.Background())
	if err != nil {
		log.Println("Ошибка при остановки сервера: ", err)
	}
	err = a.DB.Close()
	if err != nil {
		log.Println("Ошибка при закрытии соединения с БД: ", err)
	}
	err = a.Kafka.Reader.Close()
	if err != nil {
		log.Println("Ошибка при закрытии соединения с Кафкой: ", err)
	}
	return nil

}
