package app

import (
	"Level0/internal/model"
	"Level0/internal/repository"
	"log"
)

type App struct {
	DB *repository.Storage
}

func NewApp() App {
	db := repository.Init()

	return App{DB: db}
}

func (a *App) Run() {
	ch := make(chan model.Order)
	go repository.ReadFromKafka(ch)
	go func() {
		for order := range ch {
			log.Printf("Получен заказ %s из Kafka", order.OrderUID)
			a.DB.AddOrder(order)
		}
	}()

	select {}
}
