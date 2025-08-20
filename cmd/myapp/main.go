package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"level0/internal/api"
	"level0/internal/api/controllers"
	"level0/internal/app"
	"level0/internal/config"
	"level0/internal/repository"
	"level0/scripts"
)

// Загрузка переменных окружения
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// Где DEPRECATED методы я так и не понял(
func main() {
	go scripts.WriteInKafka()
	conf := config.New()
	db := repository.Init()
	cache := repository.NewLRUCache(conf.Cache.CacheSize, db)
	kafka := repository.NewKafkaReader([]string{conf.Kafka.KafkaBrokers}, conf.Kafka.KafkaTopic, conf.Kafka.KafkaConsumerGroup)
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
