package location

import (
	"time"
)

var loc *time.Location

const (
	SQLite = "2006-01-02 15:04:05"
)

func init() {
	var err error
	loc, err = time.LoadLocation("Hongkong")
	if err != nil {
		panic(err)
	}
}

// 格式化为香港时间
func FormatAsHongkong(timestamp string) string {
	// 解析时间
	utctime, err := time.Parse(SQLite, timestamp)
	if err != nil {
		return timestamp
	}

	// 时区转换
	s := utctime.Format(SQLite)
	t, _ := time.Parse(SQLite, s)
	return t.In(loc).Format(SQLite)
}
