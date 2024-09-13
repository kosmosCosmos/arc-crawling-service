package client

import (
	"github.com/kosmosCosmos/arc-crawling-service/auth"
	"github.com/kosmosCosmos/arc-crawling-service/client/doubanClient"
)

type DoubanClient struct {
	auth      auth.DoubanAuthenticator
	apiClient *doubanClient.APIClient
}

func NewDoubanClient(auth auth.DoubanAuthenticator) *DoubanClient {
	cfg := doubanClient.NewConfiguration()
	return &DoubanClient{
		auth:      auth,
		apiClient: doubanClient.NewAPIClient(cfg),
	}
}
