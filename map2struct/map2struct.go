package map2struct

import (
	"github.com/uozi-tech/cosy/internal/structcodec"
)

// WeakDecode decodes the input data to the output data with weakly typed input
func WeakDecode(input, output any) error {
	return structcodec.Decode(input, output)
}
