package main

import (
	"Level0/internal/api"
	"Level0/internal/api/controllers"
	"Level0/internal/app"
	"Level0/internal/repository"
	"Level0/scripts"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

// Загрузка переменных окружения
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	go scripts.WriteInKafka()

	db := repository.Init()
	cache := repository.NewLRUCache(100, db)
	kafka := repository.NewKafkaReader([]string{"kafka:9092"}, "orders-topic", "grp1")
	ctrl := controllers.Controller{DB: db, Cache: cache}
	mux := api.RouteController(&ctrl)

	a := app.NewApp(db, cache, kafka, mux)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := a.Run(ctx)
	if err != nil {
		return
	}

}
