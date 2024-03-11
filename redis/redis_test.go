package redis

import (
	"git.uozi.org/uozi/cosy/settings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedis(t *testing.T) {
	settings.Init("../app.ini")
	Init()

	err := Set("test", "test", 0)
	if err != nil {
		t.Error(err)
	}
	v, err := Get("test")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "test", v)

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

	keys, err := Keys("test*")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, []string{
		"test_incr",
		"test",
	}, keys)

	err = Del("test", "test_incr")
	if err != nil {
		t.Error(err)
	}
	v, _ = Get("test")
	assert.Equal(t, "", v)
	v, _ = Get("test_incr")
	assert.Equal(t, "", v)
}
