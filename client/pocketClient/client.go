package pocketClient

import (
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	MysqlClient      *xorm.Engine
	RedisClient      *redis.Client
	PocketServiceApi *PocketApiService
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Configuration) *APIClient {
	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.PocketServiceApi = (*PocketApiService)(&c.common)

	return c
}
