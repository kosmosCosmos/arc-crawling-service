package doubanClient

import (
	"xorm.io/xorm"
)

type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	MysqlConnect     *xorm.Engine
	DoubanServiceApi *DoubanServiceApiService
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.Mysql != nil {

	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.DoubanServiceApi = (*DoubanServiceApiService)(&c.common)

	return c
}
