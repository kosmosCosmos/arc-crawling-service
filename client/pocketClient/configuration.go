package pocketClient

import (
	"time"
)

// APIConfig stores API-specific configuration
type APIConfig struct {
	QuestionDetailAPI  string
	OwnerMessageAPI    string
	AlbumListApi       string
	LiveListAPI        string
	LiveDetailAPI      string
	FriendshipsURL     string
	IMServerJumpURL    string
	TeamLastMessageURL string
	TeamRoomInfoURL    string
}

// ServiceConfig stores client-specific configuration
type ServiceConfig struct {
	Header   map[string]string
	Interval time.Duration
}

// Configuration stores the configuration of the API client
type Configuration struct {
	API     APIConfig
	Service ServiceConfig
}

// DefaultHeader returns the default header configuration
func DefaultHeader() map[string]string {
	return map[string]string{"User-Agent": "PocketFans201807/7.0.2_22090903 (MuMu:Android 6.0.1;Netease V417IR release-keys)", "Host": "pocketapi.48.cn", "Connection": "Keep-Alive", "Accept-Encoding": "gzip", "appInfo": `{"IMEI":"997bfc558cc69ae6","appBuild":"22090903","appVersion":"7.0.2","deviceId":"997bfc558cc69ae6","deviceName":"MuMu","osType":"android","osVersion":"6.0.1","phoneName":"MuMu","phoneSystemVersion":"6.0.1","vendor":"Netease"}`, "Content-Type": "application/json; charset=UTF-8"}
}

// NewConfiguration returns a new Configuration object
func NewConfiguration() *Configuration {
	cfg := &Configuration{
		API: APIConfig{
			QuestionDetailAPI:  "https://pocketapi.48.cn/idolanswer/api/idolanswer/v1/question_answer/detail",
			OwnerMessageAPI:    "https://pocketapi.48.cn/im/api/v1/team/message/list/homeowner",
			AlbumListApi:       "https://pocketapi.48.cn/idolanswer/api/idolanswer/v1/user/nft/user_nft_list",
			LiveListAPI:        "https://pocketapi.48.cn/live/api/v1/live/getLiveList",
			LiveDetailAPI:      "https://pocketapi.48.cn/live/api/v1/live/getLiveOne",
			FriendshipsURL:     "https://pocketapi.48.cn/user/api/v1/friendships/friends/id",
			IMServerJumpURL:    "https://pocketapi.48.cn/im/api/v1/im/server/jump",
			TeamLastMessageURL: "https://pocketapi.48.cn/im/api/v1/team/last/message/get",
			TeamRoomInfoURL:    "https://pocketapi.48.cn/im/api/v1/im/team/room/info",
		},
		Service: ServiceConfig{
			Header:   DefaultHeader(),
			Interval: time.Hour * 24,
		},
	}

	return cfg
}

// WithInterval sets the interval for the configuration
func (c *Configuration) WithInterval(interval time.Duration) *Configuration {
	c.Service.Interval = interval
	return c
}

// WithCustomHeader sets a custom header for the configuration
func (c *Configuration) WithCustomHeader(key, value string) *Configuration {
	c.Service.Header[key] = value
	return c
}
