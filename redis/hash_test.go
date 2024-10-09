package redis

import (
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/settings"

	"testing"
)

func TestHash(t *testing.T) {
	settings.Init("../app.ini")
	Init()

	key := generateRandomKey(10)

	_, err := HSetNX(key, "field1", "value1")
	assert.NoError(t, err)

	value, err := HGet(key, "field1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	_, err = HSet(key, "field1", "value2")
	assert.NoError(t, err)

	_, err = HSetNX(key, "field2", "value1")
	assert.NoError(t, err)

	_, err = HSetNX(key, "field3", "value1")
	assert.NoError(t, err)

	all, err := HGetAll(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"field1": "value2",
		"field2": "value1",
		"field3": "value1",
	}, all)

	keys, err := HKeys(key)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"field1", "field2", "field3"}, keys)

	exists, err := HExists(key, "field1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = HExists(key, "field4")
	assert.NoError(t, err)
	assert.False(t, exists)

	length, err := HLen(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	_, err = HDel(key, "field1", "field2")
	assert.NoError(t, err)

	length, err = HLen(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// clean up
	err = Del(key)
	assert.NoError(t, err)
}
