package redis

import (
	"git.uozi.org/uozi/cosy/settings"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	settings.Init("../app.ini")
	Init()

	err := Set("test", "test", 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	v, err := Get("test")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "test", v)
	assert.LessOrEqual(t, 10*time.Second, TTL("test"))

	inc, err := Incr("test_incr")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), inc)

	incStr, err := Get("test_incr")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "1", incStr)

	inc, err = Incr("test_incr")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(2), inc)

	decr, err := Decr("test_incr")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), decr)

	keys, err := Keys("test*")
	if err != nil {
		t.Error(err)
		return
	}
	assert.ObjectsAreEqual([]string{
		"test",
		"test_incr",
	}, keys)

	err = Del("test", "test_incr")
	if err != nil {
		t.Error(err)
	}
	v, _ = Get("test")
	assert.Equal(t, "", v)
	v, _ = Get("test_incr")
	assert.Equal(t, "", v)

	err = SetEx("test", "test", 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	v, _ = Get("test")
	assert.Equal(t, "test", v)

	err = SetNx("test1", "test1", 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	v, _ = Get("test1")
	assert.Equal(t, "test1", v)

	err = SetNx("test1", "test2", 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	v, _ = Get("test1")
	assert.Equal(t, "test1", v)

}
