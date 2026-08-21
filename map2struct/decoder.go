package map2struct

import "github.com/uozi-tech/cosy/internal/structcodec"

// TypeDecoder converts an input value into one registered concrete type.
type TypeDecoder func(input any) (any, error)

// RegisterTypeDecoder registers a decoder for the concrete type represented by
// sample. Registration invalidates already compiled struct plans.
func RegisterTypeDecoder(sample any, decoder TypeDecoder) error {
	return structcodec.RegisterConverter(sample, structcodec.Converter(decoder))
}
