package auth

import (
	"fmt"
	"github.com/kosmosCosmos/arc-crawling-service/pkg/common"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
)

type ShopAuthenticatorConfig struct {
	Header   map[string]string
	Username string
	Password string
}

type ShopAuthenticator struct {
	Config ShopAuthenticatorConfig
}

func NewShopAuthenticator(config ShopAuthenticatorConfig) ShopAuthenticator {
	return ShopAuthenticator{Config: config}
}

func (sa *ShopAuthenticator) Login() (map[string]interface{}, error) {
	info, err := sa.performLogin()
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	rankInfo, err := sa.fetchRankInfo(info["cookie"].(string))
	if err != nil {
		return nil, fmt.Errorf("fetch rank info failed: %w", err)
	}

	for k, v := range rankInfo {
		info[k] = v
	}

	return info, nil
}

func (sa *ShopAuthenticator) performLogin() (map[string]interface{}, error) {
	payload := map[string]string{
		"username": sa.Config.Username,
		"password": sa.Config.Password,
	}

	resp, body, err := tools.NewRequest("POST", common.ShopLoginUrl, sa.Config.Header, payload)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	redirectURL, err := sa.extractRedirectURL(body)
	if err != nil {
		return nil, fmt.Errorf("extract redirect URL failed: %w", err)
	}

	resp, _, err = tools.NewRequest("GET", redirectURL, sa.Config.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("redirect request failed: %w", err)
	}

	cookie := sa.extractCookie(resp.Header.Values("Set-Cookie"))
	return map[string]interface{}{"cookie": cookie}, nil
}

func (sa *ShopAuthenticator) fetchRankInfo(cookie string) (map[string]interface{}, error) {
	headers := make(map[string]string)
	for k, v := range sa.Config.Header {
		headers[k] = v
	}
	headers["Cookie"] = cookie

	_, rankBody, err := tools.NewRequest("GET", common.ShopUserUrl, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch rank info request failed: %w", err)
	}

	return sa.parseRankInfo(rankBody)
}

func (sa *ShopAuthenticator) parseRankInfo(rankBody string) (map[string]interface{}, error) {
	re := regexp.MustCompile(`<span class="ic_(bs|bb|bg ?|bckg ?|bcgt ?)">(....)</span>`)
	matches := re.FindAllStringSubmatch(rankBody, -1)

	groups := map[string]string{"bs": "snh", "bb": "bej", "bg": "gnz", "bckg": "ckg", "bcgt": "cgt"}
	available := make(map[string]interface{})

	for _, match := range matches {
		if v, ok := groups[strings.TrimSpace(match[1])]; ok {
			available[v+"_sales_num"] = common.RankMap[match[2]]
		}
	}

	return available, nil
}

func (sa *ShopAuthenticator) extractCookie(cookies []string) string {
	raw := strings.Join(cookies, ";")
	cookie := strings.ReplaceAll(raw, "domain=.48.cn; path=/; HttpOnly;", "")
	return strings.ReplaceAll(cookie, " Path=/;", "")
}

func (sa *ShopAuthenticator) extractRedirectURL(responseBody string) (string, error) {
	re := regexp.MustCompile(`https:..shop.48.cn.*.m.48.cn`)
	redirectScript := re.FindString(gjson.Get(responseBody, "desc").String())
	if redirectScript == "" {
		return "", fmt.Errorf("redirect URL not found in response")
	}

	return strings.Replace(redirectScript, `" reload="1"></script>="text/javascript" src="https://m.48.cn`, "", 1), nil
}
