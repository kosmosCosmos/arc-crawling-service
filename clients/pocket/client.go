package pocket

type Client struct {
	Header map[string]string
}

func NewPocketClient(header map[string]string) *Client { return &Client{Header: header} }
