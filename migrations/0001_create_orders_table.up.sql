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