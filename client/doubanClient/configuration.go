package doubanClient

import (
	"time"
)

// APIConfig stores API-specific configuration
type APIConfig struct {
	DiscussionURL string
	TopicURL      string
}

// ClientConfig stores client-specific configuration
type ClientConfig struct {
	Header   map[string]string
	ID       string
	Interval time.Duration
}

// Configuration stores the overall configuration of the API client
type Configuration struct {
	API    APIConfig
	Client ClientConfig
}

// DefaultHeader returns the default header configuration
func DefaultHeader() map[string]string {
	return map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"accept-language": "zh-CN,zh;q=0.9",
		"cache-control":   "max-age=0",
		"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
	}
}

// NewConfiguration returns a new Configuration object with default values
func NewConfiguration() *Configuration {
	return &Configuration{
		API: APIConfig{
			DiscussionURL: "https://www.douban.com/group/%s/discussion?start=%s",
			TopicURL:      "https://www.douban.com/group/topic/%s/?start=%s",
		},
		Client: ClientConfig{
			Header:   DefaultHeader(),
			ID:       "",
			Interval: time.Hour * 24,
		},
	}
}

// WithID sets the ID for the configuration
func (c *Configuration) WithID(id string) *Configuration {
	c.Client.ID = id
	return c
}

// WithInterval sets the interval for the configuration
func (c *Configuration) WithInterval(interval time.Duration) *Configuration {
	c.Client.Interval = interval
	return c
}

// WithCustomHeader sets a custom header for the configuration
func (c *Configuration) WithCustomHeader(key, value string) *Configuration {
	c.Client.Header[key] = value
	return c
}
