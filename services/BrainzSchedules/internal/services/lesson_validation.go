package services

import (
	"brainz-api/internal/dtos"
	"fmt"
	"sort"
	"time"
)

type interval struct {
	start time.Time
	end   time.Time
	num   uint8
}

// ValidateLessonsCreate performs domain-level validation for lessons creation payload.
// It is intentionally kept independent from transport (HTTP) concerns.
func ValidateLessonsCreate(dto dtos.LessonsCreateDTO) error {
	if dto.InstitutionID <= 0 {
		return fmt.Errorf("institutionID must be positive")
	}
	if len(dto.Lessons) == 0 {
		return fmt.Errorf("lessons must not be empty")
	}

	byGroup := make(map[uint][]interval, len(dto.Lessons))
	for i, l := range dto.Lessons {
		if l.GroupID == 0 {
			return fmt.Errorf("lessons[%d].group_id must be positive", i)
		}
		byGroup[l.GroupID] = append(byGroup[l.GroupID], interval{
			start: l.StartTime,
			end:   l.EndTime,
			num:   l.Num,
		})
	}

	for groupID, items := range byGroup {
		sort.Slice(items, func(i, j int) bool { return items[i].start.Before(items[j].start) })
		for i := 1; i < len(items); i++ {
			prev := items[i-1]
			cur := items[i]
			// overlap if cur.start < prev.end (adjacent is ok).
			if cur.start.Before(prev.end) {
				return fmt.Errorf(
					"lessons overlap for group_id=%d (nums %d and %d)",
					groupID,
					prev.num,
					cur.num,
				)
			}
		}
	}

	return nil
}

