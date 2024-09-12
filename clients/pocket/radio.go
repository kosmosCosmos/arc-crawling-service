package pocket

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"sync"
	"time"
)

func (pc *Client) UpdateRadioHistory() ([]map[string]interface{}, error) {
	allLiveItems, err := pc.fetchAllLiveItems()
	if err != nil {
		return nil, fmt.Errorf("error fetching live items: %w", err)
	}

	fmt.Printf("Total live items fetched: %d\n", len(allLiveItems))

	return pc.processBroadcasts(allLiveItems)
}

func (pc *Client) fetchAllLiveItems() ([]map[string]interface{}, error) {
	var allLiveItems []map[string]interface{}
	next := int64(0)
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)

	for {
		payload := map[string]string{
			"debug":   "true",
			"groupId": "0",
			"next":    fmt.Sprintf("%d", next),
			"record":  "true",
			"teamId":  "0",
			"userId":  "0",
		}
		nextStr, liveItems, err := pc.fetchData(payload)
		if err != nil {
			return nil, fmt.Errorf("error fetching data: %w", err)
		}

		for _, item := range liveItems {
			itemTime := time.Unix(item["ctime"].(int64)/1000, 0)
			if itemTime.Before(yesterday) {
				return allLiveItems, nil
			}
			allLiveItems = append(allLiveItems, item)
		}

		if nextStr == "" {
			break
		}

		next, _ = strconv.ParseInt(nextStr, 10, 64)
	}

	return allLiveItems, nil
}

func (pc *Client) fetchData(payload map[string]string) (string, []map[string]interface{}, error) {
	_, list, err := tools.NewRequest("POST", common.LiveListAPI, pc.Header, payload)
	if err != nil {
		return "", nil, err
	}

	nextStr := gjson.Get(list, "content.next").String()
	var liveItems []map[string]interface{}

	gjson.Get(list, "content.liveList").ForEach(func(_, value gjson.Result) bool {
		item := map[string]interface{}{
			"liveId": value.Get("liveId").String(),
			"ctime":  value.Get("ctime").Int(),
		}
		liveItems = append(liveItems, item)
		return true
	})

	return nextStr, liveItems, nil
}

func (pc *Client) processBroadcasts(items []map[string]interface{}) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	broadcastsChan := make(chan map[string]interface{}, len(items))
	errChan := make(chan error, len(items))

	for _, item := range items {
		wg.Add(1)
		go func(liveId string) {
			defer wg.Done()
			broadcast, err := pc.getBroadcast(liveId)
			if err != nil {
				errChan <- fmt.Errorf("error getting broadcast for liveId %s: %w", liveId, err)
				return
			}
			broadcastsChan <- broadcast
		}(item["liveId"].(string))
	}

	go func() {
		wg.Wait()
		close(broadcastsChan)
		close(errChan)
	}()

	var broadcasts []map[string]interface{}
	for broadcast := range broadcastsChan {
		broadcasts = append(broadcasts, broadcast)
	}

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}
		return broadcasts, fmt.Errorf("some broadcasts failed to fetch")
	}

	fmt.Printf("Total broadcasts fetched: %d\n", len(broadcasts))

	return broadcasts, nil
}

func (pc *Client) getBroadcast(liveId string) (map[string]interface{}, error) {
	payload := map[string]string{
		"liveId": liveId,
	}
	_, data, err := tools.NewRequest("POST", common.LiveDetailAPI, pc.Header, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to get live details: %w", err)
	}

	content := gjson.Get(data, "content")
	broadcast := map[string]interface{}{
		"live_id":          content.Get("liveId").String(),
		"live_type":        pc.getLiveType(content.Get("liveType").Int()),
		"online_num":       content.Get("onlineNum").Int(),
		"msg_file_path":    content.Get("msgFilePath").String(),
		"owner_name":       content.Get("user.userName").String(),
		"title":            content.Get("title").String(),
		"play_stream_path": content.Get("playStreamPath").String(),
		"ctime":            pc.getFormattedTime(content.Get("ctime").Int()),
	}

	return broadcast, nil
}

func (pc *Client) getLiveType(liveType int64) string {
	switch liveType {
	case common.LiveTypeStreaming:
		return "直播"
	case common.LiveTypeRadio:
		return "电台"
	case common.LiveTypeRecording:
		return "录屏"
	default:
		return "未知"
	}
}

func (pc *Client) getFormattedTime(timestamp int64) int {
	formattedTime, err := strconv.Atoi(time.UnixMilli(timestamp).Format("20060102"))
	if err != nil {
		log.Printf("Failed to parse time: %v\n", err)
		return 0
	}
	return formattedTime
}
