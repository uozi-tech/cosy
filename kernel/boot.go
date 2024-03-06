package kernel

var async []func()

var syncs []func()

func Boot() {
	defer recovery()

	for _, v := range async {
		v()
	}

	for _, v := range syncs {
		go v()
	}
}

// RegisterAsyncFunc Register async functions, this function should be called before kernel boot.
func RegisterAsyncFunc(f ...func()) {
	async = append(async, f...)
}

// RegisterSyncsFunc Register syncs functions, this function should be called before kernel boot.
func RegisterSyncsFunc(f ...func()) {
	syncs = append(syncs, f...)
}
