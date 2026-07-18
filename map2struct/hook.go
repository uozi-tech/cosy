package map2struct

import (
	"fmt"
	"reflect"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/internal/structcodec"
)

// TypeDecoder converts an input value into one registered concrete type.
type TypeDecoder func(input any) (any, error)

// RegisterTypeDecoder registers a decoder for the concrete type represented by
// sample. Registration invalidates already compiled struct plans.
func RegisterTypeDecoder(sample any, decoder TypeDecoder) error {
	return structcodec.RegisterConverter(sample, structcodec.Converter(decoder))
}

// DecodeHookFunc is the legacy hook shape retained for source compatibility.
type DecodeHookFunc func(from, to reflect.Type, data any) (any, error)

// ToTimeHookFunc converts supported inputs to time.Time.
func ToTimeHookFunc() DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[time.Time]() {
			return data, nil
		}
		switch from.Kind() {
		case reflect.String:
			return cast.ToTimeInDefaultLocationE(data, nil)
		case reflect.Float64:
			return time.Unix(0, int64(reflect.ValueOf(data).Float())*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, reflect.ValueOf(data).Int()*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

// ToTimePtrHookFunc converts supported inputs to *time.Time.
func ToTimePtrHookFunc() DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[*time.Time]() {
			return data, nil
		}
		if from.Kind() == reflect.String && reflect.ValueOf(data).String() == "" {
			return nil, nil
		}
		converted, err := ToTimeHookFunc()(from, reflect.TypeFor[time.Time](), data)
		if err != nil {
			return nil, err
		}
		value, ok := converted.(time.Time)
		if !ok {
			return data, nil
		}
		return &value, nil
	}
}

// ToDecimalHookFunc converts string and float64 inputs to decimal.Decimal.
func ToDecimalHookFunc() DecodeHookFunc {
	return func(_, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[decimal.Decimal]() {
			return data, nil
		}
		switch value := data.(type) {
		case float64:
			return decimal.NewFromFloat(value), nil
		case string:
			if value == "" {
				return decimal.Decimal{}, nil
			}
			return decimal.NewFromString(value)
		default:
			return nil, fmt.Errorf("expected string or float64 for decimal.Decimal, got %T", data)
		}
	}
}

// ToPgDateHook converts an input to pgtype.Date.
func ToPgDateHook() DecodeHookFunc {
	return func(_, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[pgtype.Date]() {
			return data, nil
		}
		date := pgtype.Date{}
		_ = date.Set(data)
		return date, nil
	}
}

// ToPgDatePtrHook converts an input to *pgtype.Date.
func ToPgDatePtrHook() DecodeHookFunc {
	return func(_, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[*pgtype.Date]() {
			return data, nil
		}
		date := &pgtype.Date{}
		_ = date.Set(data)
		return date, nil
	}
}

// ToNullableStringHookFunc converts string inputs to null.String.
func ToNullableStringHookFunc() DecodeHookFunc {
	return func(_, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[null.String]() {
			return data, nil
		}
		value, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for null.String, got %T", data)
		}
		return null.StringFrom(value), nil
	}
}
