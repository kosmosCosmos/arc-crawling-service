package douban

type Client struct {
	Header map[string]string
	ID     string
}

func NewDoubanClient(id string, header map[string]string) *Client {
	return &Client{ID: id, Header: header}
}
