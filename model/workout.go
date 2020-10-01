package model

import (
	"time"

	"github.com/mvandergrift/energy-sdk/healthmate"
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

func NewWorkoutFromHealthmate(h healthmate.Series) (Workout, error) {
	reportedCalories := h.Workout.ManualCalories
	reportedDistance := h.Workout.ManualDistance

	if reportedCalories < 1 {
		reportedCalories = h.Workout.Calories
	}

	if reportedDistance < 1 {
		reportedDistance = h.Workout.Distance
	}

	w := Workout{
		Date:       h.Startdate.Time(),
		ActivityID: int(h.Category),
		UserID:     1,
		Duration:   h.Workout.Effduration,
		StartTime:  h.Startdate.Time(),
		EndTime:    h.Enddate.Time(),
		Calories:   reportedCalories,
		Distance:   reportedDistance,
		Comment:    "Imported from Healthmate via energy-sdk",
		ExternalID: &h.ID,
	}

	return w, nil
}
