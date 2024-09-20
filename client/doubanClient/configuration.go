package doubanClient

type MysqlConfiguration struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Configuration stores the configuration of the API client
type Configuration struct {
	Mysql         *MysqlConfiguration
	Header        map[string]string
	ID            string
	DiscussionUrl string
	TopicUrl      string
}

// NewConfiguration returns a new Configuration object
func NewConfiguration() *Configuration {
	cfg := &Configuration{
		Mysql: &MysqlConfiguration{},
		Header: map[string]string{
			"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"accept-language": "zh-CN,zh;q=0.9",
			"cache-control":   "max-age=0",
			"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		},
		ID:            "",
		DiscussionUrl: "https://www.douban.com/group/%s/discussion?start=%s",
		TopicUrl:      "https://www.douban.com/group/topic/%s/?start=%s",
	}

	return cfg
}
