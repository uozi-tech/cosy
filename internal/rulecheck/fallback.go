package rulecheck

import "github.com/go-playground/validator/v10"

func fallbackCheck(rule string) checkFn {
	return func(validate *validator.Validate, value any) checkResult {
		return runFallback(validate, value, rule)
	}
}

func runFallback(validate *validator.Validate, value any, rule string) (result checkResult) {
	if validate == nil {
		return checkFail
	}
	defer func() {
		if recover() != nil {
			result = checkFail
		}
	}()
	if validate.Var(value, rule) != nil {
		return checkFail
	}
	return checkPass
}
