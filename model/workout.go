package model

import (
	"time"
)

type Workout struct {
	ID          int
	Date        time.Time `gorm:"column:activity_date"`
	Timezone    string
	ActivityID  int
	UserID      int
	Duration    float64
	Distance    float64
	Steps       int64
	Calories    float64
	StartTime   time.Time
	EndTime     time.Time
	AttributeID int64
	ModelID     int64
	Elevation   float64
	HRAverage   int64 `gorm:"column:hr_avg"`
	HRMin       int64 `gorm:"column:hr_min"`
	HRMax       int64 `gorm:"column:hr_max"`
	Comment     string
	ExternalID  *int64
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}

func (w Workout) TableName() string {
	return "user_activity"
}
