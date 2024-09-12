package arc_crawling_service

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"log"
	"strconv"
	"strings"
	"time"
)

type DoubanClient struct {
	ID string
}

func NewDoubanClient(id string) *DoubanClient {
	return &DoubanClient{ID: id}
}

func (dc *DoubanClient) extractID(url, prefix string) string {
	return strings.Trim(strings.TrimPrefix(url, fmt.Sprintf("https://www.douban.com/%s/", prefix)), "/")
}

func (dc *DoubanClient) fetchAndParse(url string, parser func(*goquery.Document) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	_, body, err := tools.NewRequest("GET", url, common.DouBanHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request Douban page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return parser(doc)
}

func (dc *DoubanClient) FetchAndParseReplies() ([]map[string]interface{}, error) {
	var replies []map[string]interface{}
	var topicContent string
	var topicCreateTime int64

	for start := 0; ; start += 100 {
		url := fmt.Sprintf(common.TopicUrl, dc.ID, start)
		pageReplies, tContent, tCreateTime, err := dc.fetchAndParseRepliesPage(url, start == 0)
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

func (dc *DoubanClient) fetchAndParseRepliesPage(url string, updateTopicDetail bool) ([]map[string]interface{}, string, int64, error) {
	var topicContent string
	var topicCreateTime int64

	parser := func(doc *goquery.Document) ([]map[string]interface{}, error) {
		var err error

		if updateTopicDetail {
			topicContent, topicCreateTime, err = dc.extractTopicDetails(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to update topic details: %w", err)
			}
		}

		var replies []map[string]interface{}

		doc.Find(".comment-item").Each(func(i int, s *goquery.Selection) {
			reply, err := dc.extractReplyInfo(s)
			if err != nil {
				log.Printf("Failed to extract reply info: %v", err)
				return
			}
			if tools.IsRecentTime(reply["time"].(string), 0, -6, 0) {
				replies = append(replies, reply)
			}
		})

		return replies, nil
	}

	replies, err := dc.fetchAndParse(url, parser)
	if err != nil {
		return nil, "", 0, err
	}

	return replies, topicContent, topicCreateTime, nil
}

func (dc *DoubanClient) extractReplyInfo(s *goquery.Selection) (map[string]interface{}, error) {
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
		"topicId":   dc.ID,
		"username":  username,
		"userURL":   userURL,
		"content":   replyContent,
		"time":      replyTime,
		"ip":        replyIP,
		"dataCid":   dataCid,
		"likeCount": likeCount,
	}, nil
}

func (dc *DoubanClient) extractTopicDetails(doc *goquery.Document) (string, int64, error) {
	createDate := doc.Find(".create-time").Text()
	createTime, err := dateparse.ParseLocal(createDate)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse create time: %w", err)
	}

	content := strings.TrimSpace(doc.Find("#link-report > div > div").Text())

	return content, createTime.Unix(), nil
}

func (dc *DoubanClient) FetchAndParseTopics() ([]map[string]interface{}, error) {
	var allTopics []map[string]interface{}

	for page := 1; ; page++ {
		url := fmt.Sprintf(common.DiscussionUrl, dc.ID, (page-1)*50)
		topics, err := dc.parseTopics(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse topics on page %d: %w", page, err)
		}

		if len(topics) == 0 {
			break
		}

		allTopics = append(allTopics, topics...)

		log.Printf("Successfully parsed %d topics from page %d", len(topics), page)

		time.Sleep(2 * time.Minute)
	}

	return allTopics, nil
}

func (dc *DoubanClient) parseTopics(url string) ([]map[string]interface{}, error) {
	parser := func(doc *goquery.Document) ([]map[string]interface{}, error) {
		var topics []map[string]interface{}

		doc.Find("table.olt tr:not(.th)").Each(func(i int, s *goquery.Selection) {
			topic, err := dc.extractTopicInfo(s)
			if err != nil {
				log.Printf("Failed to extract topic info: %v", err)
				return
			}
			if tools.IsRecentTime(topic["lastReplyTime"].(string), 0, -6, 0) {
				topics = append(topics, topic)
			}
		})

		return topics, nil
	}

	return dc.fetchAndParse(url, parser)
}

func (dc *DoubanClient) extractTopicInfo(s *goquery.Selection) (map[string]interface{}, error) {
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
		"topicId":       dc.extractID(topicURL, "group/topic"),
		"topicUrl":      topicURL,
		"userName":      userLink.Text(),
		"userId":        dc.extractID(userURL, "people"),
		"userUrl":       userURL,
		"title":         strings.TrimSpace(topicLink.Text()),
		"topicStatus":   "unused",
		"groupId":       dc.ID,
		"replyCount":    replyCount,
		"lastReplyTime": s.Find("td.time").Text(),
	}, nil
}

func (dc *DoubanClient) GetTopicsByGroup() ([]map[string]interface{}, error) {
	topics, err := dc.FetchAndParseTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and parse topics: %w", err)
	}

	log.Printf("Successfully fetched %d topics", len(topics))

	return topics, nil
}
