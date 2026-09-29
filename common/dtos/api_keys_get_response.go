package dtos

type ApiKeysGetResponse struct {
	ApiKeys []ApiKeyShareModel `json:"keys"`
}
