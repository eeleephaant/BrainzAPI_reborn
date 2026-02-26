package models

// SessionAuthData is used for session key validation.
type SessionAuthData struct {
	ApiKeyRaw string
	IpAddress string
}
