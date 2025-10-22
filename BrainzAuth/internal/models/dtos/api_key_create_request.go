package dtos

import "github.com/gofrs/uuid/v5"

type ApiKeyCreateRequest struct {
	Title       string    `json:"title" binding:"required"`
	DeveloperId uuid.UUID `json:"developer_id" binding:"required"`
}
