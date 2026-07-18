package rulecheck

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

const maxRuleLength = 64 << 10

type checkResult uint8

const (
	checkPass checkResult = iota
	checkFail
	checkSkip
	checkUnsupported
)

type checkFn func(*validator.Validate, any) checkResult

type compiledRule struct {
	checks []checkFn
	err    error
}

var ruleCache sync.Map

func getRule(rule string) compiledRule {
	if cached, ok := ruleCache.Load(rule); ok {
		return cached.(compiledRule)
	}

	compiled := compile(rule)
	actual, _ := ruleCache.LoadOrStore(rule, compiled)
	return actual.(compiledRule)
}

func compile(rule string) compiledRule {
	if len(rule) > maxRuleLength {
		return compiledRule{err: fmt.Errorf("rule exceeds %d bytes", maxRuleLength)}
	}
	if rule == "" || rule == "-" {
		return compiledRule{}
	}

	parts := strings.Split(rule, ",")
	checks := make([]checkFn, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		token := parts[i]
		if token == "" {
			return compiledRule{err: errors.New("empty validation tag")}
		}

		name, param, hasParam := strings.Cut(token, "=")
		if name == "dive" {
			if hasParam {
				return compiledRule{err: errors.New("dive does not accept parameters")}
			}
			if i == len(parts)-1 {
				return compiledRule{err: errors.New("dive requires an element rule")}
			}
			elementRule := strings.Join(parts[i+1:], ",")
			element := compile(elementRule)
			if element.err != nil {
				return element
			}
			checks = append(checks, diveCheck(element.checks, elementRule))
			break
		}

		check, err := compileBuiltin(name, param, hasParam, token)
		if err != nil {
			return compiledRule{err: err}
		}
		checks = append(checks, check)
	}
	return compiledRule{checks: checks}
}
