package services

import "github.com/cloudwego/hertz/pkg/app/client"

type ApiKeysService struct {
	hc *client.Client
}

func NewApiKeysService(hc *client.Client) *ApiKeysService {
	return &ApiKeysService{hc: hc}
}

func (aks *ApiKeysService) CreateApiKey() {

}
func (aks *ApiKeysService) RemoveApiKey() {

}
