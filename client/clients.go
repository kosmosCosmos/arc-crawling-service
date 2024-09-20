package client

import (
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
)

type DoubanClient struct {
	ApiClient *doubanClient.APIClient
}

func NewDoubanClient(ID string, Header map[string]string, mysql doubanClient.MysqlConfiguration, redis doubanClient.RedisConfiguration) *DoubanClient {
	cfg := doubanClient.NewConfiguration()
	cfg.ID = ID
	cfg.Header = Header
	cfg.Mysql = mysql
	cfg.Redis = redis
	return &DoubanClient{
		ApiClient: doubanClient.NewAPIClient(cfg),
	}
}

func (d *DoubanClient) Hello() error {
	return d.ApiClient.DoubanServiceApi.Hello("hello")
}
