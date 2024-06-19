package sonyflake

import (
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/sony/sonyflake"
	"time"
)

var sf *sonyflake.Sonyflake

func Init() {
	var st sonyflake.Settings

	if !settings.SonyflakeSettings.StartTime.IsZero() {
		st.StartTime = settings.SonyflakeSettings.StartTime
	} else {
		st.StartTime = time.Date(2023, 3, 23, 00, 00, 00, 00, time.UTC)
	}

	if settings.SonyflakeSettings.MachineID > 0 {
		st.MachineID = func() (uint16, error) {
			return settings.SonyflakeSettings.MachineID, nil
		}
	}

	var err error
	sf, err = sonyflake.New(st)
	if err != nil {
		logger.Fatal(err)
	}
}

func NextID() uint64 {
	id, err := sf.NextID()
	if err != nil {
		return 0
	}
	return id
}
