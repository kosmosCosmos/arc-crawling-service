package doubanClient

type Config struct {
	Header map[string]string
	ID     string
}

type APIClient struct {
	Cfg                 *Config
	Common              service // Reuse a single struct instead of allocating one for each service on the heap.
	WorkspaceServiceApi *DoubanServiceApiService
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Config) *APIClient {

	c := &APIClient{}
	c.Cfg = cfg
	c.Common.client = c

	// API Services
	c.WorkspaceServiceApi = (*DoubanServiceApiService)(&c.Common)

	return c
}

type DoubanServiceApiService service
