package rulecheck

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The hand-written ASCII loop must agree with the original regexp on every
// byte and on representative mixed inputs.
func TestIsSafetyASCIIMatchesRegexp(t *testing.T) {
	reference := regexp.MustCompile(`^[a-zA-Z0-9-_./: ]*$`)
	for b := 0; b < 256; b++ {
		input := string([]byte{byte(b)})
		assert.Equal(t, reference.MatchString(input), isSafetyASCII(input), "byte %#x", b)
	}
	for _, input := range []string{"", "a_b-c.d/e:f g", "2024-01-02 10:00:00", "x\ty", "tab\t", "é", "中文", "a=b", "~", "'"} {
		assert.Equal(t, reference.MatchString(input), isSafetyASCII(input), "%q", input)
		assert.Equal(t, reference.MatchString(input) || safetyUnicodeRegex.MatchString(input), IsSafetyText(input), "%q", input)
	}
}
