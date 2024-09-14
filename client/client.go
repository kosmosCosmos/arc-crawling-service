package client

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/auth"
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"log"
	"time"
)

type DoubanClient struct {
	auth      auth.DoubanAuthenticator
	apiClient *doubanClient.APIClient
}

func NewConfiguration() *doubanClient.Config {
	return &doubanClient.Config{}
}

func NewDoubanClient(auth auth.DoubanAuthenticator, cfg doubanClient.Config) *DoubanClient {
	return &DoubanClient{
		auth:      auth,
		apiClient: doubanClient.NewAPIClient(&cfg),
	}
}

func (d *DoubanClient) FetchAndParseReplies() ([]map[string]interface{}, error) {
	var replies []map[string]interface{}
	var topicContent string
	var topicCreateTime int64

	for start := 0; ; start += 100 {
		url := fmt.Sprintf(common.TopicUrl, d.apiClient.Cfg.ID, start)
		pageReplies, tContent, tCreateTime, err := d.apiClient.WorkspaceServiceApi.FetchAndParseRepliesPage(url, start == 0)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch and parse replies on start %d: %w", start, err)
		}

		if len(pageReplies) == 0 {
			break
		}

		replies = append(replies, pageReplies...)

		if start == 0 {
			topicContent = tContent
			topicCreateTime = tCreateTime
		}

		time.Sleep(2 * time.Minute)
	}

	if len(replies) > 0 {
		replies[0]["topicContent"] = topicContent
		replies[0]["topicCreateTime"] = topicCreateTime
	}

	return replies, nil
}

func (d *DoubanClient) GetTopicsByGroup() ([]map[string]interface{}, error) {
	topics, err := d.apiClient.WorkspaceServiceApi.FetchAndParseTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and parse topics: %w", err)
	}

	log.Printf("Successfully fetched %d topics", len(topics))

	return topics, nil
}
