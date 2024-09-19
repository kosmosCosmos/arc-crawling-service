package doubanClient

type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	DoubanServiceApi *DoubanServiceApiService
}

type service struct {
	client *APIClient
}

func NewAPIClient(cfg *Configuration) *APIClient {

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.DoubanServiceApi = (*DoubanServiceApiService)(&c.common)

	return c
}
