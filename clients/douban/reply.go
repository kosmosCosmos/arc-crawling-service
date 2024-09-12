package douban

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"log"
	"strconv"
	"strings"
	"time"
)

func (dc *Client) FetchAndParseReplies() ([]map[string]interface{}, error) {
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

func (dc *Client) fetchAndParseRepliesPage(url string, updateTopicDetail bool) ([]map[string]interface{}, string, int64, error) {
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

func (dc *Client) extractReplyInfo(s *goquery.Selection) (map[string]interface{}, error) {
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
