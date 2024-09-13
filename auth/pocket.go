package auth

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
)

type PocketAuthenticatorConfig struct {
	Header map[string]string
	PocketSMSAuthenticatorConfig
	PocketLoginAuthenticatorConfig
}

type PocketSMSAuthenticatorConfig struct {
	Area   string
	Mobile string
}

type PocketLoginAuthenticatorConfig struct {
	Mobile           string
	VerificationCode string
}

type PocketAuthenticator struct{ Config PocketAuthenticatorConfig }

func NewPocketAuthenticator(config PocketAuthenticatorConfig) PocketAuthenticator {
	return PocketAuthenticator{
		Config: config,
	}
}

func (p *PocketAuthenticator) SendSMS() error {
	payload := map[string]string{
		"mobile": p.Config.PocketSMSAuthenticatorConfig.Mobile,
		"area":   p.Config.PocketSMSAuthenticatorConfig.Area,
	}
	_, _, err := tools.NewRequest("POST", common.SendSmsAPI, p.Config.Header, payload)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	return nil
}

func (p *PocketAuthenticator) LoginPocket() (string, error) {
	payload := map[string]string{
		"mobile": p.Config.PocketLoginAuthenticatorConfig.Mobile,
		"code":   p.Config.PocketLoginAuthenticatorConfig.VerificationCode,
	}
	_, body, err := tools.NewRequest("POST", common.LoginAPI, p.Config.Header, payload)
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}

	token := gjson.Get(body, "content.token").String()
	return token, nil
}
