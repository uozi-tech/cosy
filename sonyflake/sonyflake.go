package sonyflake

import (
	"time"

	"github.com/sony/sonyflake/v2"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

var sf *sonyflake.Sonyflake

// Init initializes sonyflake
func Init() {
	var st sonyflake.Settings

	if !settings.SonyflakeSettings.StartTime.IsZero() {
		st.StartTime = settings.SonyflakeSettings.StartTime
	} else {
		st.StartTime = time.Date(2023, 3, 23, 00, 00, 00, 00, time.UTC)
	}

	if settings.SonyflakeSettings.MachineID > 0 {
		st.MachineID = func() (int, error) {
			return settings.SonyflakeSettings.MachineID, nil
		}
	}

	var err error
	sf, err = sonyflake.New(st)
	if err != nil {
		logger.Fatal(err)
	}
}

// NextID generates a new sonyflake ID
func NextID() uint64 {
	id, err := sf.NextID()
	if err != nil {
		return 0
	}
	return uint64(id)
}
