package sls

// Protobuf wire encoding for SLS log types.
// Field numbers match the official SLS proto definition.

import "encoding/binary"

// Wire types
const (
	wireVarint  = 0
	wireBytes   = 2
	wireFixed32 = 5
)

func appendVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func appendTag(b []byte, field, wtype uint32) []byte {
	return appendVarint(b, uint64(field<<3|wtype))
}

func appendStringField(b []byte, field uint32, s string) []byte {
	b = appendTag(b, field, wireBytes)
	b = appendVarint(b, uint64(len(s)))
	return append(b, s...)
}

func appendOptionalStringField(b []byte, field uint32, s string) []byte {
	if s == "" {
		return b
	}
	return appendStringField(b, field, s)
}

func appendMessageField(b []byte, field uint32, msg []byte) []byte {
	b = appendTag(b, field, wireBytes)
	b = appendVarint(b, uint64(len(msg)))
	return append(b, msg...)
}

func appendUint32Field(b []byte, field, v uint32) []byte {
	b = appendTag(b, field, wireVarint)
	return appendVarint(b, uint64(v))
}

func appendFixed32Field(b []byte, field, v uint32) []byte {
	b = appendTag(b, field, wireFixed32)
	buf := [4]byte{}
	binary.LittleEndian.PutUint32(buf[:], v)
	return append(b, buf[:]...)
}

// marshalLogContent encodes a LogContent (proto fields: 1=Key, 2=Value).
func marshalLogContent(b []byte, lc *LogContent) []byte {
	b = appendStringField(b, 1, lc.Key)
	b = appendStringField(b, 2, lc.Value)
	return b
}

// marshalLog encodes a Log (proto fields: 1=Time, 2=Contents, 4=TimeNs).
func marshalLog(b []byte, l *Log) []byte {
	b = appendUint32Field(b, 1, l.Time)
	for _, c := range l.Contents {
		inner := marshalLogContent(nil, c)
		b = appendMessageField(b, 2, inner)
	}
	if l.TimeNs > 0 {
		b = appendFixed32Field(b, 4, l.TimeNs)
	}
	return b
}

// marshalLogTag encodes a LogTag (proto fields: 1=Key, 2=Value).
func marshalLogTag(b []byte, t *LogTag) []byte {
	b = appendStringField(b, 1, t.Key)
	b = appendStringField(b, 2, t.Value)
	return b
}

// MarshalLogGroup serializes a LogGroup to protobuf wire format.
// Proto fields: 1=Logs, 3=Topic, 4=Source, 6=LogTags.
func MarshalLogGroup(lg *LogGroup) []byte {
	// estimate ~100 bytes per log entry for pre-allocation
	b := make([]byte, 0, len(lg.Logs)*100+len(lg.Topic)+len(lg.Source)+64)
	for _, l := range lg.Logs {
		inner := marshalLog(nil, l)
		b = appendMessageField(b, 1, inner)
	}
	b = appendOptionalStringField(b, 3, lg.Topic)
	b = appendOptionalStringField(b, 4, lg.Source)
	for _, t := range lg.LogTags {
		inner := marshalLogTag(nil, t)
		b = appendMessageField(b, 6, inner)
	}
	return b
}
