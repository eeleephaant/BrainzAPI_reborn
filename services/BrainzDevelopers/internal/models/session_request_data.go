package models

import (
	"github.com/google/uuid"
)

type SessionRequestData struct {
	IpAddress   string
	UserAgent   string
	DeveloperId uuid.UUID
}
