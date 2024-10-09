package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/settings"
	"testing"
)

func TestSortedSet(t *testing.T) {
	settings.Init("../app.ini")
	Init()

	key := generateRandomKey(10)

	_, err := ZAdd(key, 1, "value1")
	assert.NoError(t, err)
	count, err := ZCard(key)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), count)
	}

	_, err = ZAdd(key, 2, "value2")
	assert.NoError(t, err)
	count, err = ZCount(key, "0", "1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = ZIncrBy(key, 1, "value1")
	assert.NoError(t, err)
	count, err = ZCount(key, "2", "2")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	values, err := ZRange(key, 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, values)

	_, err = ZAdd(key, 3, "value3")
	assert.NoError(t, err)

	valuesWithScores, err := ZRangeWithScores(key, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, []redis.Z{{Score: 2, Member: "value1"}, {Score: 2, Member: "value2"}, {Score: 3, Member: "value3"}}, valuesWithScores)

	values, err = ZRangeByScore(key, &redis.ZRangeBy{Min: "0", Max: "2"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, values)

	valuesWithScores, err = ZRangeByScoreWithScores(key, &redis.ZRangeBy{Min: "0", Max: "2"})
	assert.NoError(t, err)
	assert.Equal(t, []redis.Z{{Score: 2, Member: "value1"}, {Score: 2, Member: "value2"}}, valuesWithScores)

	rank, err := ZRank(key, "value2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rank)

	_, err = ZRem(key, "value1")
	assert.NoError(t, err)
	count, err = ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	_, err = ZRemRangeByRank(key, 0, 0)
	assert.NoError(t, err)
	count, err = ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = ZRemRangeByScore(key, "0", "3")
	assert.NoError(t, err)
	count, err = ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Clean up
	err = Del(key)
	assert.NoError(t, err)
}
