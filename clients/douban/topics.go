package douban

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

func (dc *Client) fetchAndParseTopics() ([]map[string]interface{}, error) {
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

func (dc *Client) GetTopicsByGroup() ([]map[string]interface{}, error) {
	topics, err := dc.fetchAndParseTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and parse topics: %w", err)
	}

	log.Printf("Successfully fetched %d topics", len(topics))

	return topics, nil
}

func (dc *Client) extractID(url, prefix string) string {
	return strings.Trim(strings.TrimPrefix(url, fmt.Sprintf("https://www.douban.com/%s/", prefix)), "/")
}

func (dc *Client) extractTopicDetails(doc *goquery.Document) (string, int64, error) {
	createDate := doc.Find(".create-time").Text()
	createTime, err := dateparse.ParseLocal(createDate)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse create time: %w", err)
	}

	content := strings.TrimSpace(doc.Find("#link-report > div > div").Text())

	return content, createTime.Unix(), nil
}

func (dc *Client) parseTopics(url string) ([]map[string]interface{}, error) {
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

func (dc *Client) extractTopicInfo(s *goquery.Selection) (map[string]interface{}, error) {
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
