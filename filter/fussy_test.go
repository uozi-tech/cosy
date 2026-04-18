package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveFuzzyClause(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(" ILIKE ?", resolveFuzzyClause("postgres"))
	assert.Equal(" LIKE ?", resolveFuzzyClause("mysql"))
	assert.Equal(" LIKE ?", resolveFuzzyClause("sqlite"))
	assert.Equal(" LIKE ?", resolveFuzzyClause(""))
	assert.Equal(" LIKE ?", resolveFuzzyClause("unknown"))
}

func TestFuzzyLikeCached(t *testing.T) {
	// Without calling model.Init the dialect name is empty, so the cached
	// clause must fall back to the non-Postgres default and stay stable.
	first := fuzzyLike()
	second := fuzzyLike()

	assert.Equal(t, " LIKE ?", first)
	assert.Equal(t, first, second)
}
