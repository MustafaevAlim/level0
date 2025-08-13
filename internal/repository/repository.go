package repository

import (
	"Level0/internal/config"
	"Level0/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
	transaction TEXT NOT NULL,
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

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) AddOrder(ctx context.Context, order model.Order) error {

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка в записи заказа в БД: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(
		ctx,
		`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES (:order_uid, :track_number, :entry, :locale, :internal_signature, :customer_id, :delivery_service, :shardkey, :sm_id, :date_created, :oof_shard)`,
		&order)
	if err != nil {
		return fmt.Errorf("ошибка в записи заказа в БД: %w", err)
	}

	order.Delivery.OrderUid = order.OrderUID
	_, err = tx.NamedExecContext(
		ctx,
		`INSERT INTO deliverys (order_uid, name, phone, zip, city, address, region, email)
		VALUES (:order_uid, :name, :phone, :zip, :city, :address, :region, :email)`,
		&order.Delivery)
	if err != nil {
		return fmt.Errorf("ошибка в записи заказа в БД: %w", err)
	}

	order.Payment.OrderUid = order.OrderUID
	_, err = tx.NamedExecContext(
		ctx,
		`INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES (:order_uid, :transaction, :request_id, :currency, :provider, :amount, :payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee)`,
		&order.Payment)
	if err != nil {
		return fmt.Errorf("ошибка в записи заказа в БД: %w", err)

	}

	for i := range order.Items {
		order.Items[i].OrderUid = order.OrderUID
		_, err = tx.NamedExecContext(
			ctx,
			`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES (:order_uid, :chrt_id, :track_number, :price, :rid, :name, :sale, :size, :total_price, :nm_id, :brand, :status)`,
			&order.Items[i])
		if err != nil {
			return fmt.Errorf("ошибка в записи заказа в БД: %w", err)
		}

	}
	log.Printf("Добавлен заказ с айди: %s", order.OrderUID)
	return tx.Commit()

}

func (s *Storage) SelectOrders(ctx context.Context, count int) ([]model.Order, error) {
	orders := make([]model.Order, 0, count)
	err := s.db.SelectContext(
		ctx,
		&orders,
		`SELECT 
        o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
        o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,

        d.name AS "deliverys.name", d.phone AS "deliverys.phone", d.zip AS "deliverys.zip",
        d.city AS "deliverys.city", d.address AS "deliverys.address", d.region AS "deliverys.region",
        d.email AS "deliverys.email",

        p.request_id AS "payments.request_id", p.transaction AS "payments.transaction",
        p.currency AS "payments.currency", p.provider AS "payments.provider", p.amount AS "payments.amount",
        p.payment_dt AS "payments.payment_dt", p.bank AS "payments.bank", p.delivery_cost AS "payments.delivery_cost",
        p.goods_total AS "payments.goods_total", p.custom_fee AS "payments.custom_fee"
		FROM orders o
		LEFT JOIN deliverys d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		ORDER BY date_created DESC LIMIT $1`,
		count,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка в чтении заказов из БД: %w", err)
	}
	for _, order := range orders {
		err = s.db.SelectContext(ctx, &order.Items, "SELECT * FROM items WHERE order_uid=$1", order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка в чтении товара заказа %s из БД: %w", order.OrderUID, err)
		}
	}
	return orders, nil
}

func (s *Storage) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	var order model.Order
	err := s.db.GetContext(
		ctx,
		&order,
		`
    	SELECT 
        o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
        o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,

        d.name AS "deliverys.name", d.phone AS "deliverys.phone", d.zip AS "deliverys.zip",
        d.city AS "deliverys.city", d.address AS "deliverys.address", d.region AS "deliverys.region",
        d.email AS "deliverys.email",

        p.request_id AS "payments.request_id", p.transaction AS "payments.transaction",
        p.currency AS "payments.currency", p.provider AS "payments.provider", p.amount AS "payments.amount",
        p.payment_dt AS "payments.payment_dt", p.bank AS "payments.bank", p.delivery_cost AS "payments.delivery_cost",
        p.goods_total AS "payments.goods_total", p.custom_fee AS "payments.custom_fee"
		FROM orders o
		LEFT JOIN deliverys d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1`,
		uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка в чтении заказа из БД: %w", err)
	}
	var items []model.Item
	err = s.db.SelectContext(ctx, &items, "SELECT * FROM items WHERE order_uid = $1", uid)
	if err != nil {
		return nil, fmt.Errorf("ошибка в чтении заказа из БД: %w", err)
	}
	order.Items = items
	return &order, nil
}
