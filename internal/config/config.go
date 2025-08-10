package config

import "os"

type RepositoryConfig struct {
	User           string
	Password       string
	RepositoryName string
}

type Config struct {
	Repository RepositoryConfig
}

func New() *Config {
	return &Config{
		Repository: RepositoryConfig{
			User:           getEnv("POSTGRES_USER", "postgres"),
			Password:       getEnv("POSTGRES_PASSWORD", "postgres"),
			RepositoryName: getEnv("POSTGRES_DB", "postgres"),
		}}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
