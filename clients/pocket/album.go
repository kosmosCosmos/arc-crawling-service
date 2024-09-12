package pocket

import (
	"encoding/json"
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"sync"
	"time"
)

func (pc *Client) UpdateAlbumHistory(rooms string) ([]map[string]interface{}, error) {
	roomArray := gjson.Get(rooms, "roomId")

	uniqueOwners := make(map[int64]bool)
	var uniqueRooms []json.RawMessage

	roomArray.ForEach(func(_, roomStr gjson.Result) bool {
		room := gjson.Parse(roomStr.String())
		ownerId := room.Get("OwnerId").Int()

		if !uniqueOwners[ownerId] {
			uniqueOwners[ownerId] = true
			uniqueRooms = append(uniqueRooms, json.RawMessage(roomStr.Raw))
		}
		return true
	})

	result, err := sjson.Set(rooms, "roomId", uniqueRooms)
	if err != nil {
		return nil, fmt.Errorf("error setting new roomId: %w", err)
	}
	OwnerIdArray := gjson.Get(result, "roomId").Array()

	var wg sync.WaitGroup
	albumsChan := make(chan []map[string]interface{}, len(OwnerIdArray))
	errChan := make(chan error, len(OwnerIdArray))

	for _, room := range OwnerIdArray {
		wg.Add(1)
		go func(room gjson.Result) {
			defer wg.Done()
			albums, err := pc.getAlbums(room)
			if err != nil {
				errChan <- fmt.Errorf("error getting albums for room: %w", err)
				return
			}
			albumsChan <- albums
		}(room)
	}

	go func() {
		wg.Wait()
		close(albumsChan)
		close(errChan)
	}()

	var allAlbums []map[string]interface{}
	for albums := range albumsChan {
		allAlbums = append(allAlbums, albums...)
	}

	for err := range errChan {
		log.Printf("Error occurred: %v", err)
	}

	fmt.Printf("Total album items fetched: %d\n", len(allAlbums))

	return allAlbums, nil
}

func (pc *Client) getAlbums(room gjson.Result) ([]map[string]interface{}, error) {
	starId := gjson.Parse(room.Value().(string)).Get("OwnerId").Int()
	ownerName := gjson.Parse(room.Value().(string)).Get("OwnerName").String()

	var allAlbums []map[string]interface{}
	page := 0
	oneDay := time.Now().Add(-24 * time.Hour)

	for {
		payload := map[string]string{
			"size":   "1000",
			"page":   fmt.Sprintf("%d", page),
			"starId": fmt.Sprintf("%d", starId),
		}

		_, body, err := tools.NewRequest("POST", common.AlbumListApi, pc.Header, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to make request for starId %d, page %d: %w", starId, page, err)
		}

		userNftListInfo := gjson.Get(body, "content.userNftListInfo")
		if len(userNftListInfo.Array()) == 0 {
			break
		}

		for _, album := range userNftListInfo.Array() {
			albumData := album.Map()
			ctime := time.Unix(albumData["createTime"].Int()/1000, 0)
			if ctime.Before(oneDay) {
				return allAlbums, nil
			}

			insert := map[string]interface{}{
				"url":        albumData["url"].String(),
				"sold":       int(albumData["sold"].Int()),
				"money":      int(albumData["money"].Int()),
				"owner_name": ownerName,
				"total":      int(albumData["total"].Int()),
				"ctime":      albumData["createTime"].Int(),
				"file_type":  pc.getFileType(albumData["sourceType"].Int()),
				"state":      pc.getState(albumData["state"].Int()),
			}
			allAlbums = append(allAlbums, insert)
		}

		page++
	}

	return allAlbums, nil
}

func (pc *Client) getFileType(sourceType int64) string {
	switch sourceType {
	case common.FileTypeImage:
		return "图片"
	case common.FileTypeAudio:
		return "音频"
	case common.FileTypeVideo:
		return "视频"
	default:
		return "未知"
	}
}

func (pc *Client) getState(state int64) string {
	switch state {
	case common.StateSale:
		return "出售中"
	case common.StateHidden:
		return "占位"
	case common.StateClosed:
		return "已结束"
	default:
		return "未知状态"
	}
}
