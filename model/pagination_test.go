package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTotalPage(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(int64(0), TotalPage(0, 10))
	assert.Equal(int64(1), TotalPage(1, 10))
	assert.Equal(int64(1), TotalPage(10, 10))
	assert.Equal(int64(2), TotalPage(11, 10))
	assert.Equal(int64(2), TotalPage(20, 10))
	assert.Equal(int64(3), TotalPage(21, 10))
	assert.Equal(int64(3), TotalPage(30, 10))
	assert.Equal(int64(4), TotalPage(31, 10))
	assert.Equal(int64(4), TotalPage(40, 10))
	assert.Equal(int64(5), TotalPage(41, 10))
	assert.Equal(int64(10), TotalPage(50, 5))
}
