package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"level0/internal/config"
	"level0/internal/model"
)

type Storage struct {
	db *sqlx.DB
}

func Init() *Storage {
	conf := config.New()
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", conf.Repository.Host, conf.Repository.User, conf.Repository.Password, conf.Repository.RepositoryName)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	return &Storage{db: db}
}

func (s *Storage) Close() error {
	return s.db.Close()

}

func (s *Storage) AddOrder(ctx context.Context, order model.Order) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка в начале транзакции: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Println("ошибка при откате транзакции: ", err)
		}
	}()

	if err := insertOrder(ctx, tx, &order); err != nil {
		return err
	}
	if err := insertDelivery(ctx, tx, &order.Delivery, order.OrderUID); err != nil {
		return err
	}
	if err := insertPayment(ctx, tx, &order.Payment, order.OrderUID); err != nil {
		return err
	}
	if err := insertItems(ctx, tx, order.Items, order.OrderUID); err != nil {
		return err
	}

	log.Printf("Добавлен заказ с айди: %s", order.OrderUID)
	return tx.Commit()
}

func insertOrder(ctx context.Context, tx *sqlx.Tx, order *model.Order) error {
	_, err := tx.NamedExecContext(ctx,
		`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service,
         shardkey, sm_id, date_created, oof_shard)
         VALUES (:order_uid, :track_number, :entry, :locale, :internal_signature, :customer_id, :delivery_service,
         :shardkey, :sm_id, :date_created, :oof_shard)`,
		order)
	if err != nil {
		return fmt.Errorf("ошибка при вставке заказа: %w", err)
	}
	return nil
}

func insertDelivery(ctx context.Context, tx *sqlx.Tx, delivery *model.Delivery, orderUID string) error {
	delivery.OrderUid = orderUID
	_, err := tx.NamedExecContext(ctx,
		`INSERT INTO deliverys (order_uid, name, phone, zip, city, address, region, email)
         VALUES (:order_uid, :name, :phone, :zip, :city, :address, :region, :email)`,
		delivery)
	if err != nil {
		return fmt.Errorf("ошибка при вставке доставки: %w", err)
	}
	return nil
}

func insertPayment(ctx context.Context, tx *sqlx.Tx, payment *model.Payment, orderUID string) error {
	payment.OrderUid = orderUID
	_, err := tx.NamedExecContext(ctx,
		`INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank,
         delivery_cost, goods_total, custom_fee)
         VALUES (:order_uid, :transaction, :request_id, :currency, :provider, :amount, :payment_dt, :bank,
         :delivery_cost, :goods_total, :custom_fee)`,
		payment)
	if err != nil {
		return fmt.Errorf("ошибка при вставке оплаты: %w", err)
	}
	return nil
}

func insertItems(ctx context.Context, tx *sqlx.Tx, items []model.Item, orderUID string) error {
	for i := range items {
		items[i].OrderUid = orderUID
	}
	_, err := tx.NamedExecContext(ctx,
		`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
         VALUES (:order_uid, :chrt_id, :track_number, :price, :rid, :name, :sale, :size, :total_price, :nm_id, :brand, :status)`,
		items)
	if err != nil {
		return fmt.Errorf("ошибка при вставке позиций: %w", err)
	}
	return nil
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
	for i := range orders {
		items, err := s.loadOrderItems(ctx, orders[i].OrderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка в чтении товара заказа %s из БД: %w", orders[i].OrderUID, err)
		}
		orders[i].Items = items

	}

	return orders, nil
}

func (s *Storage) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	var order model.Order

	if err := s.loadOrderWithDeliveryAndPayment(ctx, uid, &order); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	items, err := s.loadOrderItems(ctx, uid)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (s *Storage) loadOrderWithDeliveryAndPayment(ctx context.Context, uid string, order *model.Order) error {
	query := `
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
        WHERE o.order_uid = $1`

	return s.db.GetContext(ctx, order, query, uid)
}

func (s *Storage) loadOrderItems(ctx context.Context, uid string) ([]model.Item, error) {
	var items []model.Item
	query := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid FROM items WHERE order_uid = $1`
	err := s.db.SelectContext(ctx, &items, query, uid)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении позиций заказа: %w", err)
	}
	return items, nil
}
