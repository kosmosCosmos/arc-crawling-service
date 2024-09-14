package client

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	"github.com/kosmosCosmos/arc-crawling-service/auth"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fetchDelay     = 2 * time.Minute
	recentMonths   = -6
	topicsPerPage  = 50
	repliesPerPage = 100
)

type Config struct {
	Header map[string]string
	ID     string
}

type APIClient struct {
	Cfg                 *Config
	Common              service
	WorkspaceServiceApi *DoubanServiceApiService
	Logger              *logrus.Logger
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Config, logger *logrus.Logger) *APIClient {
	c := &APIClient{
		Cfg:    cfg,
		Logger: logger,
	}
	c.Common.client = c
	c.WorkspaceServiceApi = (*DoubanServiceApiService)(&c.Common)
	return c
}

type DoubanServiceApiService service

type DoubanClientInterface interface {
	GetTopicsByGroup(ctx context.Context) ([]map[string]interface{}, []map[string]interface{}, error)
}

type DoubanClient struct {
	auth      auth.DoubanAuthenticator
	apiClient *APIClient
}

func NewConfiguration() *Config {
	return &Config{}
}

func NewDoubanClient(auth auth.DoubanAuthenticator, cfg Config, logger *logrus.Logger) DoubanClientInterface {
	return &DoubanClient{
		auth:      auth,
		apiClient: NewAPIClient(&cfg, logger),
	}
}

func (doubanClient *DoubanClient) GetTopicsByGroup(ctx context.Context) ([]map[string]interface{}, []map[string]interface{}, error) {
	topics, err := doubanClient.apiClient.WorkspaceServiceApi.FetchAndParseTopics(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch and parse topics: %w", err)
	}

	doubanClient.apiClient.Logger.Infof("Successfully fetched %d topics", len(topics))

	var (
		replies []map[string]interface{}
		mu      sync.Mutex
		wg      sync.WaitGroup
		errChan = make(chan error, len(topics))
	)

	for _, topic := range topics {
		wg.Add(1)
		go func(t map[string]interface{}) {
			defer wg.Done()
			reply, err := doubanClient.apiClient.WorkspaceServiceApi.FetchAndParseReplies(ctx, t["topicId"].(string))
			if err != nil {
				errChan <- fmt.Errorf("failed to fetch and parse replies for topic %s: %w", t["topicId"].(string), err)
				return
			}
			mu.Lock()
			replies = append(replies, reply...)
			mu.Unlock()
		}(topic)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		doubanClient.apiClient.Logger.Error(err)
	}

	return topics, replies, nil
}

func (d *DoubanServiceApiService) FetchAndParseReplies(ctx context.Context, topicId string) ([]map[string]interface{}, error) {
	var replies []map[string]interface{}
	var topicContent string
	var topicCreateTime int64
	var err error

	for start := 0; ; start += repliesPerPage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			url := fmt.Sprintf(common.TopicUrl, topicId, start)
			var pageReplies []map[string]interface{}
			pageReplies, topicContent, topicCreateTime, err = d.FetchAndParseRepliesPage(ctx, url, start == 0)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch and parse replies on start %d: %w", start, err)
			}

			replies = append(replies, pageReplies...)

			if len(pageReplies) < repliesPerPage {
				// 如果获取的回复数量少于每页的预期数量，说明已经到达最后一页
				break
			}

			time.Sleep(fetchDelay)
		}
	}

	if len(replies) > 0 {
		replies[0]["topicContent"] = topicContent
		replies[0]["topicCreateTime"] = topicCreateTime
	}

	return replies, nil
}

func (d *DoubanServiceApiService) FetchAndParse(ctx context.Context, url string, parser func(*goquery.Document) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	_, body, err := tools.NewRequestWithContext(ctx, "GET", url, d.client.Cfg.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request Douban page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return parser(doc)
}

func (d *DoubanServiceApiService) ExtractTopicDetails(doc *goquery.Document) (string, int64, error) {
	createDate := doc.Find(".create-time").Text()
	createTime, err := dateparse.ParseLocal(createDate)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse create time: %w", err)
	}

	content := strings.TrimSpace(doc.Find("#link-report > div > div").Text())

	return content, createTime.Unix(), nil
}

func (d *DoubanServiceApiService) FetchAndParseTopics(ctx context.Context) ([]map[string]interface{}, error) {
	var allTopics []map[string]interface{}

	for page := 1; ; page++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			url := fmt.Sprintf(common.DiscussionUrl, d.client.Cfg.ID, (page-1)*topicsPerPage)
			topics, err := d.ParseTopics(ctx, url)
			if err != nil {
				return nil, fmt.Errorf("failed to parse topics on page %d: %w", page, err)
			}

			allTopics = append(allTopics, topics...)

			d.client.Logger.Infof("Successfully parsed %d topics from page %d", len(topics), page)

			if len(topics) < topicsPerPage {
				// 如果获取的主题数量少于每页的预期数量，说明已经到达最后一页
				return allTopics, nil
			}

			time.Sleep(fetchDelay)
		}
	}
}

func (d *DoubanServiceApiService) ExtractID(url, prefix string) string {
	return strings.Trim(strings.TrimPrefix(url, fmt.Sprintf("https://www.douban.com/%s/", prefix)), "/")
}

func (d *DoubanServiceApiService) ParseTopics(ctx context.Context, url string) ([]map[string]interface{}, error) {
	parser := func(doc *goquery.Document) ([]map[string]interface{}, error) {
		var topics []map[string]interface{}

		doc.Find("table.olt tr:not(.th)").Each(func(i int, s *goquery.Selection) {
			topic, err := d.ExtractTopicInfo(s)
			if err != nil {
				d.client.Logger.Errorf("Failed to extract topic info: %v", err)
				return
			}
			if tools.IsRecentTime(topic["lastReplyTime"].(string), 0, recentMonths, 0) {
				topics = append(topics, topic)
			}
		})

		return topics, nil
	}

	return d.FetchAndParse(ctx, url, parser)
}

func (d *DoubanServiceApiService) ExtractTopicInfo(s *goquery.Selection) (map[string]interface{}, error) {
	topicLink := s.Find("td.title a")
	userLink := s.Find("td:nth-child(2) a")

	topicURL, exists := topicLink.Attr("href")
	if !exists {
		return nil, fmt.Errorf("topic URL not found")
	}

	userURL, exists := userLink.Attr("href")
	if !exists {
		return nil, fmt.Errorf("user URL not found")
	}

	replyCount, _ := strconv.Atoi(s.Find("td.r-count").Text())

	return map[string]interface{}{
		"topicId":       d.ExtractID(topicURL, "group/topic"),
		"topicUrl":      topicURL,
		"userName":      userLink.Text(),
		"userId":        d.ExtractID(userURL, "people"),
		"userUrl":       userURL,
		"title":         strings.TrimSpace(topicLink.Text()),
		"topicStatus":   "unused",
		"groupId":       d.client.Cfg.ID,
		"replyCount":    replyCount,
		"lastReplyTime": s.Find("td.time").Text(),
	}, nil
}

func (d *DoubanServiceApiService) FetchAndParseRepliesPage(ctx context.Context, url string, updateTopicDetail bool) ([]map[string]interface{}, string, int64, error) {
	var topicContent string
	var topicCreateTime int64

	parser := func(doc *goquery.Document) ([]map[string]interface{}, error) {
		var err error

		if updateTopicDetail {
			topicContent, topicCreateTime, err = d.ExtractTopicDetails(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to update topic details: %w", err)
			}
		}

		var replies []map[string]interface{}

		doc.Find(".comment-item").Each(func(i int, s *goquery.Selection) {
			reply, err := d.ExtractReplyInfo(s)
			if err != nil {
				d.client.Logger.Errorf("Failed to extract reply info: %v", err)
				return
			}
			if tools.IsRecentTime(reply["time"].(string), 0, recentMonths, 0) {
				replies = append(replies, reply)
			}
		})

		return replies, nil
	}

	replies, err := d.FetchAndParse(ctx, url, parser)
	if err != nil {
		return nil, "", 0, err
	}

	return replies, topicContent, topicCreateTime, nil
}

func (d *DoubanServiceApiService) ExtractReplyInfo(s *goquery.Selection) (map[string]interface{}, error) {
	dataCid, _ := s.Attr("data-cid")
	username, _ := s.Find(".user-face img").Attr("alt")
	userURL, _ := s.Find(".user-face a").Attr("href")
	replyContent := s.Find(".reply-content").Text()
	timeIp := s.Find(".pubtime").Text()
	timeParts := strings.SplitN(timeIp, " ", 3)
	if len(timeParts) < 3 {
		return nil, fmt.Errorf("invalid time format")
	}
	replyTime := timeParts[0] + " " + timeParts[1]
	replyIP := timeParts[2]
	likeCountStr := strings.TrimSpace(strings.Trim(s.Find(".comment-vote").Text(), "赞()"))

	likeCount, _ := strconv.Atoi(likeCountStr)

	return map[string]interface{}{
		"topicId":   d.client.Cfg.ID,
		"username":  username,
		"userURL":   userURL,
		"content":   replyContent,
		"time":      replyTime,
		"ip":        replyIP,
		"dataCid":   dataCid,
		"likeCount": likeCount,
	}, nil
}
