package cosy

import (
	"testing"
)

// simple model for generic Ctx[T]
type testModel struct {
	ID int
}

// Test full lifecycle order and that nil stages are skipped.
func TestProcessChain_CreateOrModify_FullLifecycle(t *testing.T) {
	ctx := &Ctx[testModel]{}

	order := make([]string, 0, 7)
	appendOrder := func(s string) { order = append(order, s) }

	pc := NewProcessChain[testModel](ctx)
	pc.SetPrepare(func(ctx *Ctx[testModel]) { appendOrder("prepare") })
	pc.SetBeforeDecode(func(ctx *Ctx[testModel]) { appendOrder("beforeDecode") })
	pc.SetDecode(func(ctx *Ctx[testModel]) { appendOrder("decode") })
	pc.SetBeforeExecute(func(ctx *Ctx[testModel]) { appendOrder("beforeExecute") })
	pc.SetGormAction(func(ctx *Ctx[testModel]) { appendOrder("gormAction") })
	pc.SetExecuted(func(ctx *Ctx[testModel]) { appendOrder("executed") })
	pc.SetResponse(func(ctx *Ctx[testModel]) { appendOrder("response") })

	pc.CreateOrModify()

	expected := []string{"prepare", "beforeDecode", "decode", "beforeExecute", "gormAction", "executed", "response"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected length: got %d want %d, order=%v", len(order), len(expected), order)
	}
	for i, step := range expected {
		if order[i] != step {
			t.Fatalf("unexpected step at %d: got %q want %q", i, order[i], step)
		}
	}
}

// Test that aborting in the middle stops further stages.
func TestProcessChain_CreateOrModify_AbortStopsExecution(t *testing.T) {
	ctx := &Ctx[testModel]{}

	order := make([]string, 0, 7)
	appendOrder := func(s string) { order = append(order, s) }

	pc := NewProcessChain[testModel](ctx)
	pc.SetPrepare(func(ctx *Ctx[testModel]) { appendOrder("prepare") })
	pc.SetBeforeDecode(func(ctx *Ctx[testModel]) {
		appendOrder("beforeDecode")
		ctx.abort = true
	})
	pc.SetDecode(func(ctx *Ctx[testModel]) { appendOrder("decode") })
	pc.SetBeforeExecute(func(ctx *Ctx[testModel]) { appendOrder("beforeExecute") })
	pc.SetGormAction(func(ctx *Ctx[testModel]) { appendOrder("gormAction") })
	pc.SetExecuted(func(ctx *Ctx[testModel]) { appendOrder("executed") })
	pc.SetResponse(func(ctx *Ctx[testModel]) { appendOrder("response") })

	pc.CreateOrModify()

	expected := []string{"prepare", "beforeDecode"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected length after abort: got %d want %d, order=%v", len(order), len(expected), order)
	}
	for i, step := range expected {
		if order[i] != step {
			t.Fatalf("unexpected step at %d: got %q want %q", i, order[i], step)
		}
	}
}

// Test that missing (nil) stages are safely skipped and remaining run.
func TestProcessChain_CreateOrModify_SkipNilStages(t *testing.T) {
	ctx := &Ctx[testModel]{}

	order := make([]string, 0, 3)
	appendOrder := func(s string) { order = append(order, s) }

	pc := NewProcessChain(ctx)
	// Intentionally skip SetPrepare and SetDecode
	pc.SetBeforeDecode(func(ctx *Ctx[testModel]) { appendOrder("beforeDecode") })
	pc.SetBeforeExecute(func(ctx *Ctx[testModel]) { appendOrder("beforeExecute") })
	pc.SetResponse(func(ctx *Ctx[testModel]) { appendOrder("response") })

	pc.CreateOrModify()

	expected := []string{"beforeDecode", "beforeExecute", "response"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected length with nil stages: got %d want %d, order=%v", len(order), len(expected), order)
	}
	for i, step := range expected {
		if order[i] != step {
			t.Fatalf("unexpected step at %d: got %q want %q", i, order[i], step)
		}
	}
}
