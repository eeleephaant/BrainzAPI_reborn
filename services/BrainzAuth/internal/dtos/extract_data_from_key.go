package dtos

import "github.com/google/uuid"

type ExtractDataFromKey struct {
	ID     uuid.UUID
	String string
}
