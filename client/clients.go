package client

import (
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
)

type DoubanClient struct {
	ApiClient *doubanClient.APIClient
}

func NewDoubanClient(ID string, mysql *doubanClient.MysqlConfiguration) *DoubanClient {
	cfg := doubanClient.NewConfiguration()
	cfg.ID = ID
	cfg.Mysql = mysql
	return &DoubanClient{
		ApiClient: doubanClient.NewAPIClient(cfg),
	}
}

func (d *DoubanClient) UpdateTopicAndReplies() error {
	err := d.ApiClient.DoubanServiceApi.DoubanServiceApiServiceInit()
	if err != nil {
		return err
	}
	return d.ApiClient.DoubanServiceApi.UpdateTopicAndReplies()
}
