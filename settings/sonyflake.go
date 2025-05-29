package settings

import "time"

type Sonyflake struct {
	StartTime time.Time
	MachineID int
}

var SonyflakeSettings = &Sonyflake{}
