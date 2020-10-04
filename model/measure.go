package model

import "time"

type Measure struct {
	ID          int
	Date        time.Time `gorm:"column:measure_date"`
	MeasureID   int64
	UserID      int
	AttributeID int64
	ModelID     int64
	Value       float64
	ExternalID  *int64
}

func (Measure) TableName() string {
	return "user_measure"
}
