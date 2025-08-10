package repository

import (
	"Level0/internal/config"
	"Level0/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Storage struct {
	db *sqlx.DB
}

var schema = `
CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
	locale TEXT NOT NULL,
	internal_signature TEXT,
	customer_id TEXT NOT NULL,
	delivery_service TEXT NOT NULL,
	shardkey TEXT NOT NULL,
	sm_id INT NOT NULL,
	date_created TIMESTAMP WITH TIME ZONE NOT NULL,
	oof_shard TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
	id SERIAL PRIMARY KEY,
	order_uid TEXT UNIQUE NOT NULL,
	request_id TEXT,
	currency TEXT NOT NULL,
	provider TEXT NOT NULL,
	amount INT NOT NULL,
	payment_dt BIGINT NOT NULL,
	bank TEXT NOT NULL,
	delivery_cost INT NOT NULL,
	goods_total INT NOT NULL,
	custom_fee INT NOT NULL,
	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
	
);


CREATE TABLE IF NOT EXISTS deliverys (
	id SERIAL PRIMARY KEY,
	order_uid TEXT UNIQUE NOT NULL,
	name TEXT NOT NULL,
	phone TEXT NOT NULL,
	zip TEXT NOT NULL,
	city TEXT NOT NULL,
	address TEXT NOT NULL,
	region TEXT NOT NULL,
	email TEXT NOT NULL,
	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS items (
	id SERIAL PRIMARY KEY,
	order_uid TEXT NOT NULL,
	chrt_id INT NOT NULL,
	track_number TEXT NOT NULL,
	price INT NOT NULL,
	rid TEXT NOT NULL,
	name TEXT NOT NULL,
	sale INT NOT NULL,
	size TEXT NOT NULL,
	total_price INT NOT NULL,
	nm_id INT NOT NULL,
	brand TEXT NOT NULL,
	status INT NOT NULL,
	FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);

`

func Init() *Storage {
	conf := config.New()
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", conf.Repository.User, conf.Repository.Password, conf.Repository.RepositoryName)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	return &Storage{db: db}
}

func (s *Storage) AddOrder(order model.Order) error {

	tx, err := s.db.Beginx()
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Fatalln(p)
			panic(p)
		}
	}()

	_, err = tx.NamedExec(
		`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES (:order_uid, :track_number, :entry, :locale, :internal_signature, :customer_id, :delivery_service, :shardkey, :sm_id, :date_created, :oof_shard)`,
		&order)
	if err != nil {
		tx.Rollback()
		log.Fatalln(err)
		return err
	}
	order.Deliver.OrderUid = order.OrderUID
	fmt.Println(order.Deliver)
	_, err = tx.NamedExec(`INSERT INTO deliverys (order_uid, name, phone, zip, city, address, region, email)
							VALUES (:order_uid, :name, :phone, :zip, :city, :address, :region, :email)`, &order.Deliver)
	if err != nil {
		tx.Rollback()
		log.Fatalln(err)
		return err
	}
	order.Pay.OrderUid = order.OrderUID
	fmt.Println(order.Pay)
	_, err = tx.NamedExec(`INSERT INTO payments (order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
							VALUES (:order_uid, :request_id, :currency, :provider, :amount, :payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee)`, &order.Pay)
	if err != nil {
		tx.Rollback()
		log.Fatalln(err)
		return err
	}
	for i := range order.Items {

		order.Items[i].OrderUid = order.OrderUID

		_, err = tx.NamedExec(`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
							VALUES (:order_uid, :chrt_id, :track_number, :price, :rid, :name, :sale, :size, :total_price, :nm_id, :brand, :status)`, &order.Items[i])
		if err != nil {
			tx.Rollback()
			log.Fatalln(err)
			return err
		}

	}
	return tx.Commit()

}

func ReadFromKafka(ch chan model.Order) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders-topic",
	})

	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln("Ошибка при получении сообщения: ", err)
			break
		}
		var data model.Order
		err = json.Unmarshal(msg.Value, &data)

		if err != nil {
			log.Fatalln("Ошибка в десериализации сообщения: ", err)
			break
		}

		ch <- data

	}

}
