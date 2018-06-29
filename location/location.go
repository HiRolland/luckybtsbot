package location

import (
	"time"
)

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("Hongkong")
	if err != nil {
		panic(err)
	}
}

// 格式化为香港时间
func FormatAsHongkong(utctime time.Time) string {
	format := "2006-01-02 15:04:05"
	s := utctime.Format(format)
	t, _ := time.Parse(format, s)
	return t.In(loc).Format(format)
}
