package auth

import (
	"context"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/redis/go-redis/v9"

	"github.com/tidwall/gjson"
)

type Config struct {
	SendSmsAPI string
	LoginAPI   string
}

type Authenticator struct {
	phoneNumber string
	phoneArea   string
	header      map[string]string
	config      Config
	redisClient *redis.Client
}

// DefaultHeader returns the default header configuration
func DefaultHeader() map[string]string {
	return map[string]string{"User-Agent": "PocketFans201807/7.0.2_22090903 (MuMu:Android 6.0.1;Netease V417IR release-keys)", "Host": "pocketapi.48.cn", "Connection": "Keep-Alive", "Accept-Encoding": "gzip", "appInfo": `{"IMEI":"997bfc558cc69ae6","appBuild":"22090903","appVersion":"7.0.2","deviceId":"997bfc558cc69ae6","deviceName":"MuMu","osType":"android","osVersion":"6.0.1","phoneName":"MuMu","phoneSystemVersion":"6.0.1","vendor":"Netease"}`, "Content-Type": "application/json; charset=UTF-8"}
}

func NewAuthenticator(phoneNumber, phoneArea string, redisClient *redis.Client) *Authenticator {
	return &Authenticator{
		phoneNumber: phoneNumber,
		phoneArea:   phoneArea,
		header:      DefaultHeader(),
		config: Config{
			SendSmsAPI: "https://pocketapi.48.cn/user/api/v1/sms/send2",
			LoginAPI:   "https://pocketapi.48.cn/user/api/v1/login/app/mobile/code",
		},
		redisClient: redisClient,
	}
}

func (a *Authenticator) SendSMS() error {
	payload := map[string]string{
		"mobile": a.phoneNumber,
		"area":   a.phoneArea,
	}
	_, _, err := tools.NewRequest("POST", a.config.SendSmsAPI, a.header, payload)
	return err
}

func (a *Authenticator) Login(verificationCode string) error {
	payload := map[string]string{
		"mobile": a.phoneNumber,
		"code":   verificationCode,
	}
	_, body, err := tools.NewRequest("POST", a.config.LoginAPI, a.header, payload)
	if err != nil {
		return err
	}

	token := gjson.Get(body, "content.token").String()
	return a.redisClient.Set(context.Background(), "pocket_token", token, 0).Err()
}
