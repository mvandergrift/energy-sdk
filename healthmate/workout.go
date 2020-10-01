package healthmate

import (
	"log"

	"github.com/mvandergrift/energy-sdk/model"
)

type Timezone string

type WorkoutResult struct {
	Status int64 `json:"status"`
	Body   Body  `json:"body"`
}

type Body struct {
	Series []Series `json:"series"`
	More   bool     `json:"more"`
	Offset int64    `json:"offset"`
}

type Series struct {
	ID        int64     `json:"id"`
	Category  int64     `json:"category"`
	Timezone  Timezone  `json:"timezone"`
	Model     int64     `json:"model"`
	Attrib    int64     `json:"attrib"`
	Startdate Timestamp `json:"startdate"`
	Enddate   Timestamp `json:"enddate"`
	Date      string    `json:"date"`
	Deviceid  string    `json:"deviceid"`
	Workout   Data      `json:"data"`
	Modified  int64     `json:"modified"`
}

type Data struct {
	Calories       float64 `json:"calories"`
	Effduration    float64 `json:"effduration"`
	Intensity      int64   `json:"intensity"`
	ManualDistance float64 `json:"manual_distance"`
	ManualCalories float64 `json:"manual_calories"`
	Steps          int64   `json:"steps"`
	Distance       float64 `json:"distance"`
}

type Export interface {
	Display()
	ExportWorkout() (model.Workout, error)
}

func (s Series) Display() {
	log.Printf("%v:%v [%v]", s.Date, s.Category, s.Startdate)
}

func (s Series) ExportWorkout() (model.Workout, error) {
	reportedCalories := s.Workout.ManualCalories
	reportedDistance := s.Workout.ManualDistance

	if reportedCalories < 1 {
		reportedCalories = s.Workout.Calories
	}

	if reportedDistance < 1 {
		reportedDistance = s.Workout.Distance
	}

	w := model.Workout{
		Date:       s.Startdate.Time(),
		ActivityID: int(s.Category),
		UserID:     1,
		Duration:   s.Workout.Effduration,
		StartTime:  s.Startdate.Time(),
		EndTime:    s.Enddate.Time(),
		Calories:   reportedCalories,
		Distance:   reportedDistance,
		Comment:    "Imported from Healthmate via energy-sdk",
		ExternalID: &s.ID,
	}

	return w, nil
}
