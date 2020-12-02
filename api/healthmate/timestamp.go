package healthmate

import (
	"strconv"
	"time"
)

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *Timestamp) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t Timestamp) Unix() int64 {
	return time.Time(t).Unix()
}

// Time returns the JSON time as a time.Time instance in UTC
func (t Timestamp) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t Timestamp) String() string {
	return t.Time().String()
}
