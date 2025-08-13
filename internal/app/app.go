package app

import (
	"Level0/internal/model"
	"Level0/internal/repository"
	"context"
	"log"
	"net/http"
	"sync"
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

			if err := a.Kafka.Reader.CommitMessages(ctx, orderMsg.Msg); err != nil {
				log.Printf("Не удалось закоммитить offset: %v", err)
			}

		}
	}()

	server := &http.Server{
		Handler: a.Router,
		Addr:    ":8082",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP error: %v", err)
			cancel()
		}
	}()
	<-ctx.Done()

	log.Println("Завершение программы...")
	wg.Wait()

	server.Shutdown(context.Background())
	a.DB.Close()
	return nil

}
