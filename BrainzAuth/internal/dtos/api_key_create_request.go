package dtos

import "github.com/google/uuid"

type ApiKeyCreateRequest struct {
	Name        string    `json:"title,required" vd:"len($)>2 && len($)<256"`
	DeveloperId uuid.UUID `json:"developer_id,required" vd:"regexp('^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')"`
}
