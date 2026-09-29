package dtos

// ApiKeyCreateSessionRequest is the JSON body for creating an API key when the caller
// is identified by session (e.g. BrainzDevelopers: X-Session-Token). Developer ID is not sent.
type ApiKeyCreateSessionRequest struct {
	ApiKeyName  string   `json:"api_key_name,required"`
	IPWhitelist []string `json:"ip_whitelist"`
}
