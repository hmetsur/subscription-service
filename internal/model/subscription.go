package model

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"` // стоимость в рублях за месяц
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`         // первый день месяца
	EndDate     *time.Time `json:"end_date,omitempty"` // опционально, первый день месяца
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SubscriptionCreate struct {
	ServiceName string `json:"service_name"`
	Price       int64  `json:"price"`
	UserID      string `json:"user_id"`    // UUID строкой
	StartYM     string `json:"start_date"` // YYYY-MM
	EndYM       string `json:"end_date,omitempty"`
}

type SubscriptionUpdate struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int64  `json:"price,omitempty"`
	StartYM     *string `json:"start_date,omitempty"`
	EndYM       *string `json:"end_date,omitempty"`
}

type ListQuery struct {
	UserID      string
	ServiceName string
	Limit       int
	Offset      int
}

type TotalQuery struct {
	UserID      string
	ServiceName string
	From        time.Time
	To          time.Time
}

func ParseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// ParseYearMonth принимает "YYYY-MM" и возвращает первый день месяца (UTC)
func ParseYearMonth(s string) (time.Time, error) {
	return time.Parse("2006-01", s)
}
