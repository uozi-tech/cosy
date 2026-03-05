package cuid2

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	id := Generate()
	assert.Len(t, id, DefaultLength)
	assert.True(t, IsCuid(id))
}

func TestGenerateWithLength(t *testing.T) {
	for _, length := range []int{2, 10, 25, 32} {
		id := GenerateWithLength(length)
		assert.Len(t, id, length)
		assert.True(t, IsCuid(id))
	}
}

func TestGenerateWithLengthClamp(t *testing.T) {
	short := GenerateWithLength(0)
	assert.Len(t, short, MinLength)

	long := GenerateWithLength(100)
	assert.Len(t, long, MaxLength)
}

func TestGenerateStartsWithLetter(t *testing.T) {
	for range 100 {
		id := Generate()
		assert.GreaterOrEqual(t, id[0], byte('a'))
		assert.LessOrEqual(t, id[0], byte('z'))
	}
}

func TestGenerateOnlyBase36(t *testing.T) {
	for range 100 {
		id := Generate()
		for _, c := range id {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z'),
				"unexpected char %c in id %s", c, id)
		}
	}
}

func TestUniqueness(t *testing.T) {
	count := 10000
	seen := make(map[string]struct{}, count)
	for range count {
		id := Generate()
		_, exists := seen[id]
		assert.False(t, exists, "duplicate id: %s", id)
		seen[id] = struct{}{}
	}
}

func TestConcurrency(t *testing.T) {
	count := 1000
	goroutines := 10
	ids := make([]string, count*goroutines)
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for i := range count {
				ids[offset+i] = Generate()
			}
		}(g * count)
	}
	wg.Wait()

	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		assert.True(t, IsCuid(id))
		_, exists := seen[id]
		assert.False(t, exists, "duplicate id: %s", id)
		seen[id] = struct{}{}
	}
}

func TestIsCuid(t *testing.T) {
	assert.True(t, IsCuid("ab"))
	assert.True(t, IsCuid("a0123456789abcdefghijklmn"))
	assert.False(t, IsCuid(""))
	assert.False(t, IsCuid("a"))
	assert.False(t, IsCuid("1abc"))
	assert.False(t, IsCuid("Aabc"))
	assert.False(t, IsCuid("a_bc"))
	assert.False(t, IsCuid("aABC"))
}

func BenchmarkGenerate(b *testing.B) {
	for range b.N {
		Generate()
	}
}

func BenchmarkGenerateParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Generate()
		}
	})
}
