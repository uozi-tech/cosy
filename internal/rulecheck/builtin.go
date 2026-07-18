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
	safetyASCIIRegex   = regexp.MustCompile(`^[a-zA-Z0-9-_./: ]*$`)
	safetyUnicodeRegex = regexp.MustCompile(`^[\p{L}\p{N}-_.—— ]*$`)
	hostnameRegex      = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,62})*$`)
)

func compileBuiltin(name, param string, hasParam bool, token string) (checkFn, error) {
	switch name {
	case "required":
		if hasParam {
			return nil, fmt.Errorf("required does not accept parameters")
		}
		return requiredCheck, nil
	case "omitempty":
		if hasParam {
			return nil, fmt.Errorf("omitempty does not accept parameters")
		}
		return omitEmptyCheck, nil
	case "email":
		return stringCheck(validEmail, hasParam, name)
	case "url":
		return stringCheck(validURL, hasParam, name)
	case "date":
		return stringCheck(validDate, hasParam, name)
	case "safety_text":
		return stringCheck(validSafetyText, hasParam, name)
	case "hostname_port":
		return stringCheck(validHostnamePort, hasParam, name)
	case "max", "min":
		if !hasParam || param == "" {
			return nil, fmt.Errorf("%s requires a parameter", name)
		}
		limit, err := strconv.ParseFloat(param, 64)
		if err != nil || math.IsNaN(limit) || math.IsInf(limit, 0) {
			return nil, fmt.Errorf("invalid %s parameter %q", name, param)
		}
		if name == "max" {
			return func(_ *validator.Validate, value any) checkResult { return compareLimit(value, param, false) }, nil
		}
		return func(_ *validator.Validate, value any) checkResult { return compareLimit(value, param, true) }, nil
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

func compareLimit(value any, param string, minimum bool) checkResult {
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

func compareSigned(actual int64, param string, minimum bool) checkResult {
	limit, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		return checkFail
	}
	if (minimum && actual >= limit) || (!minimum && actual <= limit) {
		return checkPass
	}
	return checkFail
}

func compareUnsigned(actual uint64, param string, minimum bool) checkResult {
	limit, err := strconv.ParseUint(param, 0, 64)
	if err != nil {
		return checkFail
	}
	if (minimum && actual >= limit) || (!minimum && actual <= limit) {
		return checkPass
	}
	return checkFail
}

func compareFloat(actual float64, param string, minimum bool) checkResult {
	limit, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return checkFail
	}
	if (minimum && actual >= limit) || (!minimum && actual <= limit) {
		return checkPass
	}
	return checkFail
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return false
	}
	at := strings.LastIndexByte(address.Address, '@')
	if at <= 0 || at == len(address.Address)-1 {
		return false
	}
	domain := address.Address[at+1:]
	return strings.Contains(domain, ".") && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
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

func validDate(value string) bool {
	if len(value) != len("2006-01-02") {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validSafetyText(value string) bool {
	return safetyASCIIRegex.MatchString(value) || safetyUnicodeRegex.MatchString(value)
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
