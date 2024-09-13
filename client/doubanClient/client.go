package doubanClient

type Config struct {
	Header map[string]string
	ID     string
}

type APIClient struct {
	cfg                 *Config
	common              service // Reuse a single struct instead of allocating one for each service on the heap.
	WorkspaceServiceApi *DoubanServiceApiService
}

func NewConfiguration() *Config {
	return &Config{}
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Config) *APIClient {

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.WorkspaceServiceApi = (*DoubanServiceApiService)(&c.common)

	return c
}

type DoubanServiceApiService service
