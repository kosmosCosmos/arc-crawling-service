package doubanClient

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"log"
	"strconv"
	"strings"
	"time"
)

type DoubanApiService service

func (d *DoubanApiService) UpdateTopicAndReplies() error {
	if err := d.client.MysqlClient.Sync2(Topic{}); err != nil {
		return fmt.Errorf("failed to sync Topic table: %w", err)
	}

	if err := d.client.MysqlClient.Sync2(Reply{}); err != nil {
		return fmt.Errorf("failed to sync Reply table: %w", err)
	}

	//for page := 1; ; page++ {
	url := fmt.Sprintf(d.client.cfg.API.DiscussionURL, d.client.cfg.Client.ID, strconv.Itoa((1-1)*50))
	topics, err := d.parseTopic(url, d.client.cfg.Client.ID)
	if err != nil {
		return fmt.Errorf("failed to parse topics on page %d: %w", 1, err)
	}

	if len(topics) == 0 {
		//break
	}

	if err := d.insertTopics(topics); err != nil {
		return fmt.Errorf("failed to insert topics from page %d: %w", 1, err)
	}

	log.Printf("Successfully inserted %d topics from page %d", len(topics), 1)

	for _, topic := range topics {
		if err := d.updateRepliesByTopic(topic.TopicId); err != nil {
			log.Printf("Failed to update replies for topic %s: %v", topic.TopicId, err)
		}
	}

	time.Sleep(time.Minute * 2)
	//}

	return nil
}

func (d *DoubanApiService) updateRepliesByTopic(topicId string) error {
	for start := 0; ; start += 100 {
		url := fmt.Sprintf(d.client.cfg.API.TopicURL, topicId, strconv.Itoa(start))
		replies, err := d.parseReplies(url, topicId, start == 0)
		if err != nil {
			return fmt.Errorf("failed to parse replies on start %d: %w", start, err)
		}

		if len(replies) == 0 {
			break
		}

		if err := d.insertReplies(replies); err != nil {
			return fmt.Errorf("failed to insert replies from start %d: %w", start, err)
		}

		log.Printf("Successfully inserted %d replies from start %d for topic %s", len(replies), start, topicId)

		time.Sleep(time.Minute * 2)
	}

	affected, err := d.updateTopicStatus(topicId)
	if affected == 0 {
		return fmt.Errorf("no topic updated, possibly topic_id not found: %s, error is %v", topicId, err)
	}

	return nil
}

func (d *DoubanApiService) insertTopics(topics []*Topic) error {
	if len(topics) == 0 {
		return nil
	}

	_, err := d.client.MysqlClient.Insert(topics)
	return err
}

func (d *DoubanApiService) parseTopic(url, groupId string) ([]*Topic, error) {

	_, body, err := tools.NewRequest("GET", url, d.client.cfg.Client.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request Douban page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var topics []*Topic
	doc.Find("table.olt tr:not(.th)").Each(func(i int, s *goquery.Selection) {
		topic, err := d.extractTopicInfo(s, groupId)
		if err != nil {
			log.Printf("Failed to extract topic info: %v", err)
			return
		}
		if d.isRecentTime(topic.LastReplyTime, d.client.cfg.Client.Interval) {
			topics = append(topics, topic)
		}
	})
	return topics, nil
}

func (d *DoubanApiService) extractTopicInfo(s *goquery.Selection, groupId string) (*Topic, error) {
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

	replyCountStr := s.Find("td.r-count").Text()
	replyCountStr = strings.TrimSpace(replyCountStr)
	replyCount, err := strconv.Atoi(replyCountStr)
	if err != nil {
		replyCount = 0
	}

	lastReplyTime := strings.TrimSpace(s.Find("td.time").Text())

	return &Topic{
		TopicId:       d.extractID(topicURL, "group/topic"),
		TopicUrl:      topicURL,
		UserName:      strings.TrimSpace(userLink.Text()),
		UserId:        d.extractID(userURL, "people"),
		UserUrl:       userURL,
		Title:         strings.TrimSpace(topicLink.Text()),
		TopicStatus:   "unused",
		GroupId:       groupId,
		ReplyCount:    replyCount,
		LastReplyTime: lastReplyTime,
	}, nil
}

func (d *DoubanApiService) extractID(url, prefix string) string {
	return strings.Trim(strings.TrimPrefix(url, fmt.Sprintf("https://www.douban.com/%s/", prefix)), "/")
}

func (d *DoubanApiService) parseReplies(url, topicId string, updateTopicDetail bool) ([]*Reply, error) {
	_, body, err := tools.NewRequest("GET", url, d.client.cfg.Client.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request topic page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	if updateTopicDetail {
		if err := d.updateTopicDetails(doc, topicId); err != nil {
			return nil, fmt.Errorf("failed to update topic details: %w", err)
		}
	}

	var replies []*Reply
	doc.Find(".comment-item").Each(func(i int, s *goquery.Selection) {
		reply, err := d.extractReplyInfo(s, topicId)
		if err != nil {
			log.Printf("Failed to extract reply info: %v", err)
			return
		}
		if d.isRecentTime(reply.Time, d.client.cfg.Client.Interval) {
			replies = append(replies, reply)
		}
	})

	return replies, nil
}

func (d *DoubanApiService) updateTopicDetails(doc *goquery.Document, topicId string) error {
	createDate := strings.TrimSpace(doc.Find(".create-time").Text())
	createTime, err := dateparse.ParseIn(createDate, time.Local)
	if err != nil {
		return fmt.Errorf("failed to parse create time: %w", err)
	}

	content := strings.TrimSpace(doc.Find("#link-report > div > div").Text())

	affected, err := d.updateTopicContentAndTime(topicId, content, createTime.Unix())
	if err != nil {
		return fmt.Errorf("failed to update topic: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("no topic updated, possibly topic_id not found: %s", topicId)
	}

	return nil
}

func (d *DoubanApiService) updateTopicContentAndTime(topicId string, newContent string, newCreateTime int64) (int64, error) {
	updateTopic := &Topic{
		Content:    newContent,
		CreateTime: newCreateTime,
	}

	return d.client.MysqlClient.Where("topic_id = ?", topicId).
		Cols("content", "create_time").
		Update(updateTopic)
}

func (d *DoubanApiService) updateTopicStatus(topicId string) (int64, error) {
	updateTopic := &Topic{
		TopicStatus: "done",
	}

	return d.client.MysqlClient.Where("topic_id = ?", topicId).
		Cols("topic_status").
		Update(updateTopic)
}

func (d *DoubanApiService) insertReplies(replies []*Reply) error {
	if len(replies) == 0 {
		return nil
	}

	_, err := d.client.MysqlClient.Insert(replies)
	return err
}

func (d *DoubanApiService) extractReplyInfo(s *goquery.Selection, topicId string) (*Reply, error) {
	dataCid, exists := s.Attr("data-cid")
	if !exists {
		return nil, fmt.Errorf("data-cid attribute not found")
	}
	username := s.Find(".user-face img").AttrOr("alt", "Unknown")
	userURL := s.Find(".user-face a").AttrOr("href", "")
	if userURL == "" {
		return nil, fmt.Errorf("user URL not found")
	}

	replyContent := strings.TrimSpace(s.Find(".reply-content").Text())
	timeIp := strings.TrimSpace(s.Find(".pubtime").Text())
	timeIp = strings.ReplaceAll(timeIp, "\n", " ")

	timeParts := strings.Fields(timeIp)
	var replyTime, replyIP string
	if len(timeParts) >= 2 {
		replyTime = fmt.Sprintf("%s %s", timeParts[0], timeParts[1])
		if len(timeParts) >= 3 {
			replyIP = timeParts[2]
		}
	} else {
		replyTime = timeIp
	}
	likeCountStr := s.Find(".reply-opts .comment-vote .count").Text()
	likeCountStr = strings.TrimSpace(likeCountStr)
	if likeCountStr == "" {
		likeCountStr = "0"
	}
	likeCount, err := strconv.Atoi(likeCountStr)
	if err != nil {
		likeCount = 0
	}

	return &Reply{
		TopicId:   topicId,
		Username:  username,
		UserURL:   userURL,
		Content:   replyContent,
		Time:      replyTime,
		IP:        replyIP,
		DataCid:   dataCid,
		LikeCount: likeCount,
	}, nil
}

func (d *DoubanApiService) isRecentTime(dateStr string, interval time.Duration) bool {
	parsedTime, _ := d.parseDateTime(dateStr)
	isRecent := parsedTime.Before(time.Now().Add(interval))
	log.Printf("Time: %s, is recent: %t", parsedTime.Format("2006-01-02 15:04"), isRecent)
	return isRecent
}

func (d *DoubanApiService) parseDateTime(dateStr string) (time.Time, error) {
	fullFormat := "2006-01-02 15:04:05"
	t, err := time.Parse(fullFormat, dateStr)
	if err == nil {
		return t, nil
	}

	shortFormat := "01-02 15:04"
	t, err = time.Parse(shortFormat, dateStr)
	if err == nil {
		currentYear := time.Now().Year()
		t = t.AddDate(currentYear, 0, 0)
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", dateStr)
}
