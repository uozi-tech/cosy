package cosy

import "github.com/0xJacky/cosy/model"

func getHook[T any]() func(core *Ctx[T]) {
	resolved := model.GetResolvedModel[T]()

	return func(core *Ctx[T]) {
		var preloads []string

		for _, field := range resolved.OrderedFields {
			dir := field.CosyTag.GetItem()
			switch dir {
			case Preload:
				preloads = append(preloads, field.Name)
			}
		}
		if len(preloads) > 0 {
			core.SetPreloads(preloads...)
		}
	}
}

func getListHook[T any]() func(core *Ctx[T]) {
	resolved := model.GetResolvedModel[T]()

	return func(core *Ctx[T]) {
		var (
			in      []string
			eq      []string
			fussy   []string
			orIn    []string
			orEq    []string
			orFussy []string
			preload []string
			search  []string
		)

		for _, field := range resolved.OrderedFields {
			dirs := field.CosyTag.GetList()

			for _, dir := range dirs {
				switch dir {
				case In:
					in = append(in, field.JsonTag)
				case Equal:
					eq = append(eq, field.JsonTag)
				case Fussy:
					fussy = append(fussy, field.JsonTag)
				case OrIn:
					orIn = append(orIn, field.JsonTag)
				case OrEqual:
					orEq = append(orEq, field.JsonTag)
				case OrFussy:
					orFussy = append(orFussy, field.JsonTag)
				case Preload:
					preload = append(preload, field.Name)
				case Search:
					search = append(search, field.JsonTag)
				}
			}
		}

		if len(in) > 0 {
			core.SetIn(in...)
		}
		if len(eq) > 0 {
			core.SetEqual(eq...)
		}
		if len(fussy) > 0 {
			core.SetFussy(fussy...)
		}
		if len(orIn) > 0 {
			core.SetOrIn(orIn...)
		}
		if len(orEq) > 0 {
			core.SetOrEqual(orEq...)
		}
		if len(orFussy) > 0 {
			core.SetOrFussy(orFussy...)
		}
		if len(preload) > 0 {
			core.SetPreloads(preload...)
		}
		if len(search) > 0 {
			core.SetSearchFussyKeys(search...)
		}
	}
}
