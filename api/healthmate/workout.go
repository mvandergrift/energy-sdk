package healthmate

import (
	"fmt"

	"github.com/mvandergrift/energy-sdk/model"
)

type WorkoutResult struct {
	Status int64 `json:"status"`
	Body   struct {
		Series []Series `json:"series"`
		More   bool     `json:"more"`
		Offset int64    `json:"offset"`
	} `json:"body"`
}

type Series struct {
	ID        int64     `json:"id"`
	Category  int64     `json:"category"`
	Timezone  string    `json:"timezone"`
	Model     int64     `json:"model"`
	Attrib    int64     `json:"attrib"`
	Startdate Timestamp `json:"startdate"`
	Enddate   Timestamp `json:"enddate"`
	Date      string    `json:"date"`
	Deviceid  string    `json:"deviceid"`
	Workout   struct {
		Calories       float64 `json:"calories"`
		Effduration    float64 `json:"effduration"`
		Intensity      int64   `json:"intensity"`
		ManualDistance float64 `json:"manual_distance"`
		ManualCalories float64 `json:"manual_calories"`
		Steps          int64   `json:"steps"`
		Distance       float64 `json:"distance"`
		Elevation      float64 `json:"elevation"`
		HRAverage      int64   `json:"hr_average"`
		HRMin          int64   `json:"hr_min"`
		HRMax          int64   `json:"hr_max"`
	} `json:"data"`
	Modified int64 `json:"modified"`
}

func (s Series) String() string {
	return fmt.Sprintf("Healthmate - workout\t%v\t%v\t%v", s.Date, s.Startdate.Time().Local(), s.Category)
}

func (s Series) Export() (interface{}, error) {
	reportedCalories := s.Workout.Calories
	reportedDistance := s.Workout.Distance

	if s.Attrib == 7 && s.Workout.ManualCalories > 5 {
		reportedCalories = s.Workout.ManualCalories
	}

	if s.Attrib == 7 && s.Workout.ManualDistance > 5 {
		reportedDistance = s.Workout.ManualDistance
	}

	w := model.Workout{
		Date:        s.Startdate.Time(),
		Timezone:    s.Timezone,
		ActivityID:  int(s.Category),
		UserID:      1,
		Duration:    s.Workout.Effduration,
		StartTime:   s.Startdate.Time(),
		EndTime:     s.Enddate.Time(),
		Calories:    reportedCalories,
		Distance:    reportedDistance,
		Steps:       s.Workout.Steps,
		Elevation:   s.Workout.Elevation,
		HRAverage:   s.Workout.HRAverage,
		HRMax:       s.Workout.HRMax,
		HRMin:       s.Workout.HRMin,
		AttributeID: s.Attrib,
		ModelID:     s.Model,
		Comment:     "Imported from energy-sdk v0.01",
		ExternalID:  &s.ID,
	}

	return w, nil
}
