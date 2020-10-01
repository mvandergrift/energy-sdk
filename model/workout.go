package model

import (
	"time"
)

type Workout struct {
	ID         int
	Date       time.Time `gorm:"column:activity_date"`
	ActivityID int
	UserID     int
	Duration   float64
	Distance   float64
	Calories   float64
	StartTime  time.Time
	EndTime    time.Time
	Comment    string
	ExternalID *int64
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time
}

func (w Workout) TableName() string {
	return "user_activity"
}
