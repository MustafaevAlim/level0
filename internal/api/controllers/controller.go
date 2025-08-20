package controllers

import "level0/internal/repository"

// В принципе можно использовать только кеш, бд на будущее пусть будет
type Controller struct {
	DB    *repository.Storage
	Cache *repository.LRUcache
}
