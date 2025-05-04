package kernel

import "context"

var async []func()

var syncs []func(context.Context)

// Boot the kernel
func Boot(ctx context.Context) {
	defer recovery()

	for _, v := range async {
		v()
	}

	for _, v := range syncs {
		go v(ctx)
	}
}

// RegisterInitFunc Register init functions, this function should be called before kernel boot.
func RegisterInitFunc(f ...func()) {
	async = append(async, f...)
}

// RegisterGoroutine Register syncs functions, this function should be called before kernel boot.
func RegisterGoroutine(f ...func(context.Context)) {
	syncs = append(syncs, f...)
}
