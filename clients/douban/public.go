package douban

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/kosmosCosmos/arc-golang-toolkit/tools"
	"strings"
)

func (dc *Client) fetchAndParse(url string, parser func(*goquery.Document) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	_, body, err := tools.NewRequest("GET", url, dc.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request Douban page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return parser(doc)
}
