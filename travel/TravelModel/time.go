package TravelModel

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type CustomTime time.Time

const timeFormat = "2006-01-02 15:04:05"
const timezone = "Asia/Shanghai"

func (t CustomTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}
func (t *CustomTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = CustomTime(now)
	return
}

func (t CustomTime) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t CustomTime) local() time.Time {
	loc, _ := time.LoadLocation(timezone)
	return time.Time(t).In(loc)
}
func (t CustomTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(t)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

func (t *CustomTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = CustomTime(value)
		return nil
	}
	return fmt.Errorf("cant not convert %v to timestamp", v)
}
