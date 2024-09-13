package pocket

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"log"
	"sync"
)

func (pc *Client) UpdateChannelInfo() ([]map[string]interface{}, error) {
	_, body, err := tools.NewRequest("POST", common.FriendshipsURL, pc.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request friendships: %w", err)
	}

	friends := gjson.Get(body, "content.data").Array()

	var wg sync.WaitGroup
	channelsChan := make(chan []map[string]interface{}, len(friends))
	errChan := make(chan error, len(friends))

	for _, friend := range friends {
		wg.Add(1)
		go func(f gjson.Result) {
			defer wg.Done()
			channelInfo, err := pc.getChannel(f)
			if err != nil {
				errChan <- fmt.Errorf("failed to get channel info: %w", err)
				return
			}
			channelsChan <- channelInfo
		}(friend)
	}

	go func() {
		wg.Wait()
		close(channelsChan)
		close(errChan)
	}()

	var allChannels []map[string]interface{}
	for channels := range channelsChan {
		allChannels = append(allChannels, channels...)
	}

	for err := range errChan {
		log.Printf("Error occurred: %v", err)
	}

	fmt.Printf("Total channel items fetched: %d\n", len(allChannels))

	return allChannels, nil
}

func (pc *Client) getServerIDForStarID(starID int) (int, error) {
	payload := map[string]string{
		"targetType": "1",
		"tabId":      "0",
		"starId":     fmt.Sprintf("%d", starID),
	}
	_, body, err := tools.NewRequest("POST", common.IMServerJumpURL, pc.Header, payload)
	if err != nil {
		return 0, fmt.Errorf("failed to request server ID: %w", err)
	}
	serverID := int(gjson.Parse(body).Get("content.serverId").Int())
	return serverID, nil
}

func (pc *Client) getLastMsgList(serverID int) ([]gjson.Result, error) {
	channelPayload := map[string]string{
		"serverId": fmt.Sprintf("%d", serverID),
	}
	_, channelBody, err := tools.NewRequest("POST", common.TeamLastMessageURL, pc.Header, channelPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to get last message list: %w", err)
	}
	lastMsgList := gjson.Parse(channelBody).Get("content.lastMsgList").Array()
	return lastMsgList, nil
}

func (pc *Client) getChannelInfo(channelID int) (gjson.Result, error) {
	infoPayload := map[string]string{
		"channelId": fmt.Sprintf("%d", channelID),
	}
	_, infoBody, err := tools.NewRequest("POST", common.TeamRoomInfoURL, pc.Header, infoPayload)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to get channel info: %w", err)
	}
	channelInfo := gjson.Parse(infoBody).Get("content.channelInfo")
	return channelInfo, nil
}

func (pc *Client) getChannel(friend gjson.Result) ([]map[string]interface{}, error) {
	var channels []map[string]interface{}

	serverID, err := pc.getServerIDForStarID(int(friend.Int()))
	if err != nil {
		return nil, fmt.Errorf("failed to get server ID: %w", err)
	}

	lastMsgList, err := pc.getLastMsgList(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last message list: %w", err)
	}

	for _, channel := range lastMsgList {
		channelID := int(channel.Get("channelId").Int())
		channelInfo, channelErr := pc.getChannelInfo(channelID)
		if channelErr != nil {
			log.Printf("Failed to get channel info: %v\n", channelErr)
			continue
		}

		if channelInfo.Get("functionType").String() == "CHAT_CHANNEL" {
			insert := map[string]interface{}{
				"ChannelName": channelInfo.Get("channelName").String(),
				"ChannelId":   channelInfo.Get("channelId").Int(),
				"OwnerId":     channelInfo.Get("ownerId").Int(),
				"ServerId":    serverID,
				"OwnerName":   channelInfo.Get("ownerName").String(),
			}
			channels = append(channels, insert)
		}
	}

	return channels, nil
}
