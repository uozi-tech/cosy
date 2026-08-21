package rulecheck

import (
	"fmt"
	"math"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

var (
	// safetyUnicodeRegex is the second alternative of the safety_text rule;
	// the first, `^[a-zA-Z0-9-_./: ]*$`, is evaluated by isSafetyASCII.
	safetyUnicodeRegex = regexp.MustCompile(`^[\p{L}\p{N}-_.—— ]*$`)
	// hostnameRegex is validator's hostnameRegexStringRFC1123: labels must end
	// with an alphanumeric character.
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	// emailRegex is validator's emailRegexString, so the compiled "email" rule
	// accepts exactly what the validator fallback accepts.
	emailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
)

func compileBuiltin(name, param string, hasParam bool, token string, nullable bool) (checkFn, error) {
	switch name {
	case "required":
		if hasParam {
			return nil, fmt.Errorf("required does not accept parameters")
		}
		if nullable {
			return requiredNullableCheck, nil
		}
		return requiredCheck, nil
	case "omitempty":
		if hasParam {
			return nil, fmt.Errorf("omitempty does not accept parameters")
		}
		if nullable {
			return omitEmptyNullableCheck, nil
		}
		return omitEmptyCheck, nil
	case "omitzero":
		// validator's omitzero tests the zero value even for interface-wrapped
		// elements, so it never takes the nullable variant.
		if hasParam {
			return nil, fmt.Errorf("omitzero does not accept parameters")
		}
		return omitEmptyCheck, nil
	case "omitnil":
		if hasParam {
			return nil, fmt.Errorf("omitnil does not accept parameters")
		}
		return omitNilCheck, nil
	case "email":
		return stringCheck(validEmail, hasParam, name)
	case "url":
		return stringCheck(validURL, hasParam, name)
	case "date":
		return stringCheck(IsDate, hasParam, name)
	case "safety_text":
		return stringCheck(IsSafetyText, hasParam, name)
	case "hostname_port":
		return stringCheck(validHostnamePort, hasParam, name)
	case "max", "min":
		if !hasParam || param == "" {
			return nil, fmt.Errorf("%s requires a parameter", name)
		}
		limit, err := parseLimit(param)
		if err != nil {
			return nil, fmt.Errorf("invalid %s parameter %q", name, param)
		}
		minimum := name == "min"
		return func(_ *validator.Validate, value any) checkResult { return compareLimit(value, limit, minimum) }, nil
	case "oneof":
		if !hasParam || strings.TrimSpace(param) == "" {
			return nil, fmt.Errorf("oneof requires at least one value")
		}
		values, err := parseOneOf(param)
		if err != nil {
			return nil, err
		}
		return func(_ *validator.Validate, value any) checkResult {
			actual, ok := oneOfValue(value)
			if !ok {
				return checkUnsupported
			}
			for _, expected := range values {
				if actual == expected {
					return checkPass
				}
			}
			return checkFail
		}, nil
	default:
		return fallbackCheck(token), nil
	}
}

func stringCheck(validate func(string) bool, hasParam bool, name string) (checkFn, error) {
	if hasParam {
		return nil, fmt.Errorf("%s does not accept parameters", name)
	}
	return func(_ *validator.Validate, value any) checkResult {
		text, ok := value.(string)
		if !ok {
			return checkUnsupported
		}
		if validate(text) {
			return checkPass
		}
		return checkFail
	}, nil
}

func requiredCheck(_ *validator.Validate, value any) checkResult {
	switch value := value.(type) {
	case nil:
		return checkFail
	case string:
		if value == "" {
			return checkFail
		}
	case bool:
		if !value {
			return checkFail
		}
	case int:
		if value == 0 {
			return checkFail
		}
	case int8:
		if value == 0 {
			return checkFail
		}
	case int16:
		if value == 0 {
			return checkFail
		}
	case int32:
		if value == 0 {
			return checkFail
		}
	case int64:
		if value == 0 {
			return checkFail
		}
	case uint:
		if value == 0 {
			return checkFail
		}
	case uint8:
		if value == 0 {
			return checkFail
		}
	case uint16:
		if value == 0 {
			return checkFail
		}
	case uint32:
		if value == 0 {
			return checkFail
		}
	case uint64:
		if value == 0 {
			return checkFail
		}
	case uintptr:
		if value == 0 {
			return checkFail
		}
	case float32:
		if value == 0 {
			return checkFail
		}
	case float64:
		if value == 0 {
			return checkFail
		}
	case []any:
		// validator's required checks nil-ness, not collection length.
		if value == nil {
			return checkFail
		}
	case []string:
		if value == nil {
			return checkFail
		}
	case map[string]any:
		if value == nil {
			return checkFail
		}
	default:
		return checkUnsupported
	}
	return checkPass
}

// isNilValue reports whether value is nil or a typed nil collection, the only
// states validator treats as "empty" for interface-wrapped values.
func isNilValue(value any) bool {
	switch value := value.(type) {
	case nil:
		return true
	case []any:
		return value == nil
	case []string:
		return value == nil
	case map[string]any:
		return value == nil
	}
	return false
}

func requiredNullableCheck(_ *validator.Validate, value any) checkResult {
	if isNilValue(value) {
		return checkFail
	}
	return checkPass
}

func omitEmptyNullableCheck(_ *validator.Validate, value any) checkResult {
	if isNilValue(value) {
		return checkSkip
	}
	return checkPass
}

// omitNilCheck skips typed nil collections only: validator does not treat an
// untyped nil map value as nil here (the following checks still run and fail
// through the fallback), and typed nil pointers reach the fallback unchanged.
func omitNilCheck(_ *validator.Validate, value any) checkResult {
	switch value := value.(type) {
	case []any:
		if value == nil {
			return checkSkip
		}
	case []string:
		if value == nil {
			return checkSkip
		}
	case map[string]any:
		if value == nil {
			return checkSkip
		}
	}
	return checkPass
}

func omitEmptyCheck(validate *validator.Validate, value any) checkResult {
	switch requiredCheck(validate, value) {
	case checkFail:
		return checkSkip
	case checkUnsupported:
		return checkUnsupported
	default:
		return checkPass
	}
}

// limitParam holds the min/max parameter parsed once per rule. Each numeric
// path keeps its own parse result so the per-kind semantics (base-0 integer
// literals, sign and overflow handling) stay exactly as validator's.
type limitParam struct {
	signed     int64
	signedOK   bool
	unsigned   uint64
	unsignedOK bool
	float      float64
}

// parseLimit parses a min/max parameter once. Integer literals use base 0 as
// validator does (so 0x10 and 010 are accepted); the float form is derived
// from the integer parse when the literal is not a valid float.
func parseLimit(param string) (limitParam, error) {
	var limit limitParam
	var err error
	limit.signed, err = strconv.ParseInt(param, 0, 64)
	limit.signedOK = err == nil
	limit.unsigned, err = strconv.ParseUint(param, 0, 64)
	limit.unsignedOK = err == nil
	float, err := strconv.ParseFloat(param, 64)
	switch {
	case err == nil && !math.IsNaN(float) && !math.IsInf(float, 0):
		limit.float = float
	case limit.signedOK:
		limit.float = float64(limit.signed)
	case limit.unsignedOK:
		limit.float = float64(limit.unsigned)
	default:
		return limitParam{}, fmt.Errorf("not a finite number")
	}
	return limit, nil
}

func compareLimit(value any, param limitParam, minimum bool) checkResult {
	switch value := value.(type) {
	case string:
		return compareSigned(int64(utf8.RuneCountInString(value)), param, minimum)
	case []any:
		return compareSigned(int64(len(value)), param, minimum)
	case []string:
		return compareSigned(int64(len(value)), param, minimum)
	case map[string]any:
		return compareSigned(int64(len(value)), param, minimum)
	case int:
		return compareSigned(int64(value), param, minimum)
	case int8:
		return compareSigned(int64(value), param, minimum)
	case int16:
		return compareSigned(int64(value), param, minimum)
	case int32:
		return compareSigned(int64(value), param, minimum)
	case int64:
		return compareSigned(value, param, minimum)
	case uint:
		return compareUnsigned(uint64(value), param, minimum)
	case uint8:
		return compareUnsigned(uint64(value), param, minimum)
	case uint16:
		return compareUnsigned(uint64(value), param, minimum)
	case uint32:
		return compareUnsigned(uint64(value), param, minimum)
	case uint64:
		return compareUnsigned(value, param, minimum)
	case uintptr:
		return compareUnsigned(uint64(value), param, minimum)
	case float32:
		return compareFloat(float64(value), param, minimum)
	case float64:
		return compareFloat(value, param, minimum)
	default:
		return checkUnsupported
	}
}

func compareSigned(actual int64, param limitParam, minimum bool) checkResult {
	if !param.signedOK {
		return checkFail
	}
	return compareOrdered(actual, param.signed, minimum)
}

func compareUnsigned(actual uint64, param limitParam, minimum bool) checkResult {
	if !param.unsignedOK {
		return checkFail
	}
	return compareOrdered(actual, param.unsigned, minimum)
}

func compareFloat(actual float64, param limitParam, minimum bool) checkResult {
	return compareOrdered(actual, param.float, minimum)
}

func compareOrdered[T int64 | uint64 | float64](actual, limit T, minimum bool) checkResult {
	if (minimum && actual >= limit) || (!minimum && actual <= limit) {
		return checkPass
	}
	return checkFail
}

// validEmail mirrors validator's isEmail: net/mail must parse the address and
// validator's email regex must match it.
func validEmail(value string) bool {
	if _, err := mail.ParseAddress(value); err != nil {
		return false
	}
	return emailRegex.MatchString(value)
}

func validURL(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := url.Parse(strings.ToLower(value))
	if err != nil || parsed.Scheme == "" {
		return false
	}
	if parsed.Scheme == "file" {
		return parsed.Path != "" && parsed.Path != "/"
	}
	return parsed.Host != "" || parsed.Fragment != "" || parsed.Opaque != ""
}

// IsDate reports whether value is a valid YYYY-MM-DD date. It backs both the
// compiled "date" rule and valid.IsDate.
func IsDate(value string) bool {
	if len(value) != len("2006-01-02") {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// IsSafetyText reports whether value only contains characters allowed by the
// "safety_text" rule. It backs both the compiled rule and valid.SafetyText.
func IsSafetyText(value string) bool {
	return isSafetyASCII(value) || safetyUnicodeRegex.MatchString(value)
}

// isSafetyASCII is `^[a-zA-Z0-9-_./: ]*$` as a byte loop: the common case
// (ASCII identifiers, paths, timestamps) never reaches the regexp engine.
func isSafetyASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		switch c := value[i]; {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-', c == '_', c == '.', c == '/', c == ':', c == ' ':
		default:
			return false
		}
	}
	return true
}

func validHostnamePort(value string) bool {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return false
	}
	portNumber, err := strconv.ParseInt(port, 10, 32)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return false
	}
	return host == "" || hostnameRegex.MatchString(host)
}

func parseOneOf(param string) ([]string, error) {
	values := make([]string, 0, 4)
	for len(param) > 0 {
		param = strings.TrimLeft(param, " ")
		if param == "" {
			break
		}
		if param[0] != '\'' {
			end := strings.IndexByte(param, ' ')
			if end < 0 {
				end = len(param)
			}
			values = append(values, param[:end])
			param = param[end:]
			continue
		}
		param = param[1:]
		end := strings.IndexByte(param, '\'')
		if end < 0 {
			return nil, fmt.Errorf("unclosed oneof parameter")
		}
		values = append(values, param[:end])
		param = param[end+1:]
		if param != "" && param[0] != ' ' {
			return nil, fmt.Errorf("malformed oneof parameter")
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("oneof requires at least one value")
	}
	return values, nil
}

func oneOfValue(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case int:
		return strconv.FormatInt(int64(value), 10), true
	case int8:
		return strconv.FormatInt(int64(value), 10), true
	case int16:
		return strconv.FormatInt(int64(value), 10), true
	case int32:
		return strconv.FormatInt(int64(value), 10), true
	case int64:
		return strconv.FormatInt(value, 10), true
	case uint:
		return strconv.FormatUint(uint64(value), 10), true
	case uint8:
		return strconv.FormatUint(uint64(value), 10), true
	case uint16:
		return strconv.FormatUint(uint64(value), 10), true
	case uint32:
		return strconv.FormatUint(uint64(value), 10), true
	case uint64:
		return strconv.FormatUint(value, 10), true
	default:
		return "", false
	}
}
