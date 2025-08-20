package config

import (
	"log"
	"os"
	"strconv"
)

type RepositoryConfig struct {
	User           string
	Password       string
	RepositoryName string
	Host           string
}

type KafkaConfig struct {
	KafkaBrokers       string // TODO: сделай чтоб много брокеров было
	KafkaTopic         string
	KafkaConsumerGroup string
}

type CacheConfig struct {
	CacheSize int
}

type ServerConfig struct {
	Port string
}

type Config struct {
	Repository RepositoryConfig
	Kafka      KafkaConfig
	Cache      CacheConfig
	Server     ServerConfig
}

func New() *Config {
	return &Config{
		Repository: RepositoryConfig{
			User:           getEnv("POSTGRES_USER", "postgres"),
			Password:       getEnv("POSTGRES_PASSWORD", "postgres"),
			RepositoryName: getEnv("POSTGRES_DB", "postgres"),
			Host:           getEnv("DB_HOST", "postgres"),
		},
		Kafka: KafkaConfig{
			KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
			KafkaTopic:         getEnv("KAFKA_TOPIC", "orders-topic"),
			KafkaConsumerGroup: getEnv("KAFKA_GROUP", "grp1"),
		},
		Cache: CacheConfig{
			CacheSize: getIntEnv("CACHE_SIZE", 100),
		},
		Server: ServerConfig{
			Port: getEnv("HTTP_PORT", "8082"),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaulValue int) int {
	n, err := strconv.Atoi(getEnv(key, "100"))
	if err != nil {
		log.Println("Ошибка в получении int значения из .env")
		return defaulValue
	}
	return n
}
