package kernel

var async []func()

var syncs []func()

// Boot the kernel
func Boot() {
	defer recovery()

	for _, v := range async {
		v()
	}

	for _, v := range syncs {
		go v()
	}
}

// RegisterInitFunc Register init functions, this function should be called before kernel boot.
func RegisterInitFunc(f ...func()) {
	async = append(async, f...)
}

// RegisterGoroutine Register syncs functions, this function should be called before kernel boot.
func RegisterGoroutine(f ...func()) {
	syncs = append(syncs, f...)
}
