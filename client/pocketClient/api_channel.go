package pocketClient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"log"
	"sync"
)

type PocketApiService service

func (p *PocketApiService) UpdateChannelInfo() error {
	_, body, err := tools.NewRequest("POST", p.client.cfg.API.FriendshipsURL, p.client.cfg.Service.Header, nil)
	if err != nil {
		return fmt.Errorf("http获取失败: %w", err)
	}

	friends := gjson.Get(body, "content.data").Array()
	var wg sync.WaitGroup
	for _, friend := range friends {
		wg.Add(1)
		go func(f gjson.Result) {
			defer wg.Done()
			p.getChannel(f)
		}(friend)
	}
	wg.Wait()

	room := make(map[string]interface{})
	roomStr, err := p.client.RedisClient.LRange(context.Background(), "channels_list", 0, -1).Result()
	if err != nil {
		return fmt.Errorf("redis获取room失败: %w", err)
	}

	room["roomId"] = roomStr
	roomJson, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("marshal room失败: %w", err)
	}

	err = p.client.RedisClient.Set(context.Background(), "channels", string(roomJson), -1).Err()
	if err != nil {
		return fmt.Errorf("设置channels失败: %w", err)
	}

	err = p.client.RedisClient.Del(context.Background(), "channels_list").Err()
	if err != nil {
		return fmt.Errorf("删除channels_list失败: %w", err)
	}

	return nil
}

func (p *PocketApiService) getServerIDForStarID(starID int) (int, error) {
	payload := map[string]interface{}{
		"targetType": 1,
		"tabId":      0,
		"starId":     starID,
	}
	_, body, err := tools.NewRequest("POST", p.client.cfg.API.IMServerJumpURL, p.client.cfg.Service.Header, payload)
	if err != nil {
		return 0, err
	}
	serverID := int(gjson.Parse(body).Get("content.serverId").Int())
	return serverID, nil
}

func (p *PocketApiService) getLastMsgList(serverID int) ([]gjson.Result, error) {
	channelPayload := map[string]interface{}{
		"serverId": serverID,
	}
	_, channelBody, err := tools.NewRequest("POST", p.client.cfg.API.TeamLastMessageURL, p.client.cfg.Service.Header, channelPayload)
	if err != nil {
		return nil, err
	}
	lastMsgList := gjson.Parse(channelBody).Get("content.lastMsgList").Array()
	return lastMsgList, nil
}

func (p *PocketApiService) getChannelInfo(channelID int) (gjson.Result, error) {
	infoPayload := map[string]interface{}{
		"channelId": channelID,
	}
	_, infoBody, err := tools.NewRequest("POST", p.client.cfg.API.TeamRoomInfoURL, p.client.cfg.Service.Header, infoPayload)
	if err != nil {
		return gjson.Result{}, err
	}
	channelInfo := gjson.Parse(infoBody).Get("content.channelInfo")
	return channelInfo, nil
}

func (p *PocketApiService) getChannel(friend gjson.Result) {
	serverID, err := p.getServerIDForStarID(int(friend.Int()))
	if err != nil {
		log.Printf("Failed to get server ID: %v\n", err)
		return
	}

	lastMsgList, err := p.getLastMsgList(serverID)
	if err != nil {
		log.Printf("Failed to get last message list: %v\n", err)
		return
	}

	for _, channel := range lastMsgList {
		channelID := int(channel.Get("channelId").Int())
		channelInfo, channelErr := p.getChannelInfo(channelID)
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
			channelJson, channelJsonErr := json.Marshal(insert)
			if channelJsonErr != nil {
				log.Printf("Failed to marshal channel: %v\n", channelJsonErr)
				continue
			}
			err := p.client.RedisClient.RPush(context.Background(), "channels_list", string(channelJson)).Err()
			if err != nil {
				log.Printf("Failed to push channel to Redis: %v\n", err)
			}
		}
	}
}
