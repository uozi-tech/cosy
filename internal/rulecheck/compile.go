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
	return compileRule(rule, false)
}

// compileRule compiles a rule string. nullable selects validator's semantics
// for interface-wrapped values (the elements of a []any under dive): required
// fails only on nil and omitempty/omitzero skip only nil, instead of the
// zero-value semantics used for top-level map values.
func compileRule(rule string, nullable bool) compiledRule {
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
			element := compileRule(elementRule, false)
			if element.err != nil {
				return element
			}
			nullableElement := compileRule(elementRule, true)
			if nullableElement.err != nil {
				return nullableElement
			}
			checks = append(checks, diveCheck(element.checks, nullableElement.checks, elementRule))
			break
		}

		if needsValidator(name, token) {
			checks = append(checks, fallbackCheck(token))
			continue
		}
		check, err := compileBuiltin(name, param, hasParam, token, nullable)
		if err != nil {
			return compiledRule{err: err}
		}
		checks = append(checks, check)
	}
	return compiledRule{checks: checks}
}

// needsValidator reports whether a token must be evaluated by validator itself:
// the tag was overridden through Override, or the token uses validator syntax
// the built-in compiler does not model ('|' alternatives, 0x2C / 0x7C escapes).
func needsValidator(name, token string) bool {
	return isOverridden(name) ||
		strings.ContainsRune(token, '|') ||
		strings.Contains(token, "0x2C") ||
		strings.Contains(token, "0x7C")
}
