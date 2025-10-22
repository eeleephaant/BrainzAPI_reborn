package dtos

import "time"

type DailyUsage struct {
	Date  time.Time `json:"date"`
	Count uint32    `json:"count"`
}
