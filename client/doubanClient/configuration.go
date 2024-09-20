package doubanClient

type MysqlConfiguration struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type RedisConfiguration struct {
	Host     string
	Port     int
	Password string
	Database int
}

// Configuration stores the configuration of the API client
type Configuration struct {
	Mysql  MysqlConfiguration
	Redis  RedisConfiguration
	Header map[string]string
	ID     string
}

// NewConfiguration returns a new Configuration object
func NewConfiguration() *Configuration {
	cfg := &Configuration{
		Mysql:  MysqlConfiguration{},
		Redis:  RedisConfiguration{},
		Header: make(map[string]string),
		ID:     "",
	}

	return cfg
}
