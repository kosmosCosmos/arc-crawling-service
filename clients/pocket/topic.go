package pocket

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"log"
	"sync"
	"time"
)

func (pc *Client) UpdateTopicHistory(rooms string, msgType string) ([]map[string]interface{}, []map[string]interface{}, error) {
	roomArray := gjson.Get(rooms, "roomId").Array()
	if len(roomArray) == 0 {
		return nil, nil, fmt.Errorf("no rooms found")
	}

	var wg sync.WaitGroup
	qsChan := make(chan []map[string]interface{}, len(roomArray))
	contentChan := make(chan []map[string]interface{}, len(roomArray))
	errChan := make(chan error, len(roomArray))

	for _, room := range roomArray {
		wg.Add(1)
		go func(room gjson.Result) {
			defer wg.Done()
			qsDocuments, contentDocuments, err := pc.processRoom(room, msgType)
			if err != nil {
				errChan <- fmt.Errorf("error processing room: %w", err)
				return
			}
			qsChan <- qsDocuments
			contentChan <- contentDocuments
		}(room)
	}

	go func() {
		wg.Wait()
		close(qsChan)
		close(contentChan)
		close(errChan)
	}()

	var allQsDocuments, allContentDocuments []map[string]interface{}
	for qs := range qsChan {
		allQsDocuments = append(allQsDocuments, qs...)
	}
	for content := range contentChan {
		allContentDocuments = append(allContentDocuments, content...)
	}

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return nil, nil, fmt.Errorf("encountered %d errors during processing: %v", len(errors), errors)
	}

	fmt.Printf("Total QS documents: %d, Total Content documents: %d\n", len(allQsDocuments), len(allContentDocuments))

	return allQsDocuments, allContentDocuments, nil
}

func (pc *Client) processRoom(room gjson.Result, msgType string) ([]map[string]interface{}, []map[string]interface{}, error) {
	serverId := gjson.Parse(room.String()).Get("ServerId").Int()
	channelId := gjson.Parse(room.String()).Get("ChannelId").Int()
	OwnerName := gjson.Parse(room.String()).Get("OwnerName").String()

	return pc.collectDocuments(serverId, channelId, OwnerName, msgType)
}

func (pc *Client) collectDocuments(serverId, channelId int64, OwnerName string, mode string) ([]map[string]interface{}, []map[string]interface{}, error) {
	var qsDocuments, contentDocuments []map[string]interface{}
	nextTime := time.Now().UnixMilli()
	startTime := time.Now().Add(-24 * time.Hour)

	for {
		messages, err := pc.fetchMessages(serverId, channelId, nextTime)
		if err != nil {
			return nil, nil, err
		}

		if len(messages) == 0 {
			break
		}

		for _, info := range messages {
			msgTime := time.UnixMilli(info.Get("msgTime").Int())
			if startTime.After(msgTime) {
				return qsDocuments, contentDocuments, nil
			}

			if info.Get("msgType").String() == "FLIPCARD" && (mode == "answer" || mode == "all") {
				document, err := pc.processFlipCard(info)
				if err != nil {
					log.Printf("Error processing flipcard: %v", err)
					continue
				}
				qsDocuments = append(qsDocuments, document)
			} else if info.Get("msgType").String() == "TEXT" && (mode == "chat" || mode == "all") {
				document, err := pc.processText(info, OwnerName)
				if err != nil {
					log.Printf("Error processing text: %v", err)
					continue
				}
				contentDocuments = append(contentDocuments, document)
			}
		}

		nextTime = messages[len(messages)-1].Get("msgTime").Int()

		if startTime.After(time.UnixMilli(nextTime)) {
			break
		}
	}

	return qsDocuments, contentDocuments, nil
}

func (pc *Client) fetchMessages(serverId, channelId, nextTime int64) ([]gjson.Result, error) {
	payload := map[string]string{
		"serverId":  fmt.Sprintf("%d", serverId),
		"channelId": fmt.Sprintf("%d", channelId),
		"nextTime":  fmt.Sprintf("%d", nextTime),
		"limit":     "30",
	}
	_, body, err := tools.NewRequest("POST", common.OwnerMessageAPI, pc.Header, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return gjson.Parse(body).Get("content.message").Array(), nil
}

func (pc *Client) processFlipCard(info gjson.Result) (map[string]interface{}, error) {
	flipCardInfo := gjson.Parse(info.Get("bodys").String()).Get("filpCardInfo")
	payload := map[string]string{
		"answerId":   flipCardInfo.Get("answerId").String(),
		"questionId": flipCardInfo.Get("questionId").String(),
	}
	_, questionBody, err := tools.NewRequest("POST", common.QuestionDetailAPI, pc.Header, payload)
	if err != nil {
		return nil, fmt.Errorf("error fetching question details: %w", err)
	}
	qs := gjson.Parse(questionBody).Get("content")
	return map[string]interface{}{
		"question_id":   flipCardInfo.Get("questionId").String(),
		"user_name":     qs.Get("userName").String(),
		"question":      qs.Get("question").String(),
		"answer":        qs.Get("answer").String(),
		"question_time": qs.Get("questionTime").Int(),
		"answer_time":   qs.Get("answerTime").Int(),
		"member_name":   qs.Get("memberName").String(),
		"cost":          qs.Get("cost").Int(),
	}, nil
}

func (pc *Client) processText(info gjson.Result, OwnerName string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"msg_id":      info.Get("msgIdServer").String(),
		"msg_time":    info.Get("msgTime").Int(),
		"member_name": OwnerName,
		"content":     info.Get("bodys").String(),
	}, nil
}
