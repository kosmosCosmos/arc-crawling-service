package client

import (
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
	"github.com/kosmosCosmos/arc-crawling-service/client/pocketClient"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

type DoubanClient struct {
	ApiClient *doubanClient.APIClient
}

type ChannelPocketClient struct {
	ApiClient *pocketClient.APIClient
}

func NewDoubanClient(ID string) *DoubanClient {
	return &DoubanClient{
		ApiClient: doubanClient.NewAPIClient(doubanClient.NewConfiguration().WithID(ID)),
	}
}

func (d *DoubanClient) UpdateTopicAndReplies(conn *xorm.Engine) error {
	d.ApiClient.MysqlClient = conn
	return d.ApiClient.DoubanServiceApi.UpdateTopicAndReplies()
}

func NewChannelClient() *ChannelPocketClient {
	return &ChannelPocketClient{
		ApiClient: pocketClient.NewAPIClient(pocketClient.NewConfiguration()),
	}
}

func (c *ChannelPocketClient) UpdateChannel(mysqlConn *xorm.Engine, redisConn *redis.Client) error {
	c.ApiClient.MysqlClient = mysqlConn
	c.ApiClient.RedisClient = redisConn
	return c.ApiClient.PocketServiceApi.UpdateChannelInfo()
}
