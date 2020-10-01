package healthmate

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
