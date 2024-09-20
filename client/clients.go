package client

import (
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
	"xorm.io/xorm"
)

type DoubanClient struct {
	ApiClient *doubanClient.APIClient
}

func NewDoubanClient(ID string) *DoubanClient {
	cfg := doubanClient.NewConfiguration()
	cfg.ID = ID
	return &DoubanClient{
		ApiClient: doubanClient.NewAPIClient(cfg),
	}
}

func (d *DoubanClient) UpdateTopicAndReplies(conn *xorm.Engine) error {
	d.ApiClient.MysqlConnect = conn
	return d.ApiClient.DoubanServiceApi.UpdateTopicAndReplies()
}
