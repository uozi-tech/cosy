package map2struct

import (
	"github.com/mitchellh/mapstructure"
)

// WeakDecode decodes the input data to the output data with weakly typed input
func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToDecimalHookFunc(),
			ToNullableStringHookFunc(),
			ToTimeHookFunc(),
			ToTimePtrHookFunc(),
			ToPgDateHook(),
			ToPgDatePtrHook(),
		),
		TagName: "json",
		Squash:  true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
