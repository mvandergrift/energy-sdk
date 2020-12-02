package healthmate

import (
	"fmt"
	"math"

	"github.com/mvandergrift/energy-sdk/model"
)

type MeasureResult struct {
	Status int64 `json:"status"`
	Body   struct {
		Updatetime  Timestamp `json:"updatetime"`
		Timezone    string    `json:"timezone"`
		Measuregrps []struct {
			Grpid        int64     `json:"grpid"`
			Attrib       int64     `json:"attrib"`
			Date         Timestamp `json:"date"`
			Created      Timestamp `json:"created"`
			Category     int64     `json:"category"`
			Deviceid     string    `json:"deviceid"`
			HashDeviceid string    `json:"hash_deviceid"`
			Measures     []Measure `json:"measures"`
			Comment      string    `json:"comment"`
		} `json:"measuregrps"`
	} `json:"body"`
}

type Measure struct {
	Value       int64 `json:"value"`
	Type        int64 `json:"type"`
	Unit        int64 `json:"unit"`
	Date        Timestamp
	Grpid       int64
	AttributeID int64
}

func (s Measure) String() string {
	return fmt.Sprintf("Healthmate - measure\t%v\t%v\t%v\t[%v]", s.Date.Time().Local(), s.Type, s.Value, s.Unit)
}

func (s Measure) Export() (interface{}, error) {
	// generate unique ExternalID based on groupID and typeID
	cantorPairing := (s.Grpid+s.Type)*(s.Grpid+s.Type+1)/2 + s.Type

	w := model.Measure{
		UserID:      1,
		Date:        s.Date.Time(),
		ExternalID:  &cantorPairing,
		MeasureID:   s.Type,
		Value:       float64(s.Value) * math.Pow10(int(s.Unit)),
		AttributeID: s.AttributeID,
	}

	return w, nil
}
