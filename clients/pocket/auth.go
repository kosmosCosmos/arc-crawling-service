package pocket

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"strconv"
)

func (pc *Client) SendSMS(phoneNumber string, phoneArea int) error {
	payload := map[string]string{
		"mobile": phoneNumber,
		"area":   strconv.Itoa(phoneArea),
	}
	_, _, err := tools.NewRequest("POST", common.SendSmsAPI, pc.Header, payload)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	return nil
}

func (pc *Client) Login(phoneNumber string, verificationCode int) (string, error) {
	payload := map[string]string{
		"mobile": phoneNumber,
		"code":   strconv.Itoa(verificationCode),
	}
	_, body, err := tools.NewRequest("POST", common.LoginAPI, pc.Header, payload)
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}

	token := gjson.Parse(body).Get("content.token").String()

	return token, nil
}
