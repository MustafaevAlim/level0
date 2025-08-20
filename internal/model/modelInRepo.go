package model

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderMsg struct {
	Msg   kafka.Message
	Order Order
}

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid" validate:"required"`
	TrackNumber       string    `json:"track_number" db:"track_number" validate:"required"`
	Entry             string    `json:"entry" db:"entry" validate:"required"`
	Locale            string    `json:"locale" db:"locale" validate:"required,len=2"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature"`
	CustomerId        string    `json:"customer_id" db:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" validate:"required"`
	ShardKey          string    `json:"shardkey" db:"shardkey" validate:"required"`
	SmID              int       `json:"sm_id" db:"sm_id" validate:"required"`
	CreatedAt         time.Time `json:"date_created" db:"date_created" validate:"required,notzero"`
	OofShard          string    `json:"oof_shard" db:"oof_shard"  validate:"required"`
	Delivery          Delivery  `json:"delivery" db:"deliverys" validate:"required,dive"`
	Payment           Payment   `json:"payment" db:"payments" validate:"required,dive"`
	Items             []Item    `json:"items" db:"items" validate:"required,dive,min=1"`
}

type Delivery struct {
	ID       int    `json:"id" db:"id"`
	OrderUid string `json:"order_uid" db:"order_uid"`
	Name     string `json:"name" db:"name" validate:"required"`
	Phone    string `json:"phone" db:"phone" validate:"required,e164" `
	Zip      string `json:"zip" db:"zip" validate:"required,numeric"`
	City     string `json:"city" db:"city" validate:"required"`
	Address  string `json:"address" db:"address" validate:"required"`
	Region   string `json:"region" db:"region" validate:"required"`
	Email    string `json:"email" db:"email" validate:"required,email"`
}

type Payment struct {
	ID           int    `json:"id" db:"id"`
	OrderUid     string `json:"order_uid" db:"order_uid"`
	Transaction  string `json:"transaction" db:"transaction" validate:"required"`
	RequestId    string `json:"request_id" db:"request_id"`
	Currency     string `json:"currency" db:"currency" validate:"required,len=3"`
	Provider     string `json:"provider" db:"provider" validate:"required"`
	Amount       int    `json:"amount" db:"amount" validate:"gte=0"`
	PaymentDt    int    `json:"payment_dt" db:"payment_dt" validate:"required"`
	Bank         string `json:"bank" db:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ID          int    `db:"id"`
	OrderUid    string `db:"order_uid"`
	ChrtId      int    `json:"chrt_id" db:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" db:"track_number" validate:"required"`
	Price       int    `json:"price" db:"price" validate:"gte=0"`
	Rid         string `json:"rid" db:"rid" validate:"required"`
	Name        string `json:"name" db:"name" validate:"required"`
	Sale        int    `json:"sale" db:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size" db:"size"`
	TotalPrice  int    `json:"total_price" db:"total_price" validate:"gte=0"`
	NmId        int    `json:"nm_id" db:"nm_id" validate:"required"`
	Brand       string `json:"brand" db:"brand" validate:"required"`
	Status      int    `json:"status" db:"status" validate:"required"`
}
