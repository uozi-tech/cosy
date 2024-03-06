package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCosyTag(t *testing.T) {
	assert := assert.New(t)
	tag := "all:max=100;add:required;list:fussy;update:omitempty;item:preload"
	c := NewCosyTag(tag)
	assert.Equal("required,max=100", c.GetAdd())
	assert.Equal("omitempty,max=100", c.GetUpdate())
	assert.Equal("fussy", c.GetList()[0])
	assert.Equal("preload", c.GetItem())

	tag = "add:required;list:in;update:omitempty"
	c = NewCosyTag(tag)
	assert.Equal("required", c.GetAdd())
	assert.Equal("omitempty", c.GetUpdate())
	assert.Equal("in", c.GetList()[0])

	tag = "add:required;list:preload;update:omitempty;item:preload"
	c = NewCosyTag(tag)
	assert.Equal("required", c.GetAdd())
	assert.Equal("omitempty", c.GetUpdate())
	assert.Equal("preload", c.GetList()[0])

	tag = "add:required;list:in,search;update:omitempty;item:preload"
	c = NewCosyTag(tag)
	assert.Equal("required", c.GetAdd())
	assert.Equal("omitempty", c.GetUpdate())
	assert.Equal("in", c.GetList()[0])
	assert.Equal("search", c.GetList()[1])
	assert.Equal("preload", c.GetItem())
}
