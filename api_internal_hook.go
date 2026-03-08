package cosy

import (
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/model"
)

func populateColumnMapping[T any](core *Ctx[T], resolved *model.ResolvedModel) {
	for _, field := range resolved.OrderedFields {
		if field.JsonTag != "" && field.JsonTag != "-" && field.DBName != "" {
			core.columnMapping[field.JsonTag] = field.DBName
		}

		if field.CosyTag.GetJson() != "" && field.DBName != "" {
			core.columnMapping[field.CosyTag.GetJson()] = field.DBName
		}
	}
}

func getHook[T any]() func(core *Ctx[T]) {
	resolved := model.GetResolvedModel[T]()

	return func(core *Ctx[T]) {
		populateColumnMapping(core, resolved)

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
		populateColumnMapping(core, resolved)

		for _, field := range resolved.OrderedFields {
			dirs := field.CosyTag.GetList()

			for _, dir := range dirs {
				switch dir {
				case In:
					core.SetIn(field.JsonTag)
				case Equal:
					core.SetEqual(field.JsonTag)
				case Fussy:
					core.SetFussy(field.JsonTag)
				case OrIn:
					core.SetOrIn(field.JsonTag)
				case OrEqual:
					core.SetOrEqual(field.JsonTag)
				case OrFussy:
					core.SetOrFussy(field.JsonTag)
				case Preload:
					core.SetPreloads(field.Name)
				case Search:
					core.SetSearchFussyKeys(field.JsonTag)
				case Between:
					core.SetBetween(field.JsonTag)
				default:
					core.SetCustomFilter(field.JsonTag, dir)
				}
			}

			if field.CosyTag.GetUnique() {
				core.SetUnique(field.JsonTag)
			}
		}
	}
}

func createHook[T any]() func(core *Ctx[T]) {
	resolved := model.GetResolvedModel[T]()
	return func(core *Ctx[T]) {
		populateColumnMapping(core, resolved)

		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetAdd()
			if dirs == "" {
				continue
			}
			key := field.JsonTag
			// like password field we don't need to response it to client,
			// but we need to validate it
			if key == "-" {
				if field.CosyTag.GetJson() != "" {
					key = field.CosyTag.GetJson()
				} else {
					continue
				}
			}

			validMap[key] = dirs

			if field.Unique || field.CosyTag.GetUnique() {
				core.SetUnique(key)
			}
		}
		core.SetValidRules(validMap)
	}
}

func modifyHook[T any]() func(core *Ctx[T]) {
	resolved := model.GetResolvedModel[T]()
	return func(core *Ctx[T]) {
		populateColumnMapping(core, resolved)

		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetUpdate()
			if dirs == "" {
				continue
			}
			key := field.JsonTag
			// like password field, we don't need to response it to the client,
			// but we need to validate it
			if key == "-" {
				if field.CosyTag.GetJson() != "" {
					key = field.CosyTag.GetJson()
				} else {
					continue
				}
			}

			validMap[key] = dirs

			if field.Unique || field.CosyTag.GetUnique() {
				core.SetUnique(key)
			}
		}
		core.SetValidRules(validMap)
	}
}
