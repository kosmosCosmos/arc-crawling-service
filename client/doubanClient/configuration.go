package doubanClient

// Configuration stores the configuration of the API client
type Configuration struct {
	Header map[string]string
	ID     string
}

// NewConfiguration returns a new Configuration object
func NewConfiguration() *Configuration {
	cfg := &Configuration{
		Header: make(map[string]string),
		ID:     "",
	}

	return cfg
}
