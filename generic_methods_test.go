package cosy

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type genericMethodModel struct {
	Name string
}

type genericMethodResult struct {
	Label string
}

func TestSetTransformerInfersConcreteResultType(t *testing.T) {
	ctx := &Ctx[genericMethodModel]{}

	returned := ctx.SetTransformer(func(value *genericMethodModel) genericMethodResult {
		return genericMethodResult{Label: value.Name}
	})

	require.Same(t, ctx, returned)
	require.Equal(t,
		genericMethodResult{Label: "cosy"},
		ctx.transformer(&genericMethodModel{Name: "cosy"}),
	)
}

func TestSetScanInfersConcreteResultType(t *testing.T) {
	ctx := &Ctx[genericMethodModel]{}
	want := []genericMethodResult{{Label: "cosy"}}

	returned := ctx.SetScan(func(_ *gorm.DB) []genericMethodResult {
		return want
	})

	require.Same(t, ctx, returned)
	require.Equal(t, want, ctx.scan(nil))
}
