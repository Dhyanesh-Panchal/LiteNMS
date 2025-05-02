package containers

import (
	"path"
	"strconv"
	"time"
)

type PolledDataPoint struct {
	Timestamp uint32 `json:"timestamp" msgpack:"timestamp"`

	CounterId uint16 `json:"counter_id" msgpack:"counter_id"`

	ObjectId uint32 `json:"object_id" msgpack:"object_id"`

	Value interface{} `json:"value" msgpack:"value"`
}

type Date struct {
	Day int

	Month int

	Year int
}

func (date Date) Format() string {

	return path.Join(strconv.Itoa(date.Year), strconv.Itoa(date.Month), strconv.Itoa(date.Day))

}

func UnixToDate(unix uint32) Date {

	t := time.Unix(int64(unix), 0)

	return Date{

		Day: t.Day(),

		Month: int(t.Month()),

		Year: t.Year(),
	}

}

//func TimeToDate(t time.Time) Date {
//
//	return Date{
//		Day:   t.Day(),
//		Month: int(t.Month()),
//		Year:  t.Year(),
//	}
//
//}
