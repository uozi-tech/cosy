package settings

import "time"

type Sonyflake struct {
	StartTime time.Time
	MachineID uint16
}

var SonyflakeSettings = &Sonyflake{}
