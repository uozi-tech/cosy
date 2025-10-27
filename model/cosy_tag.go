package model

import (
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

type CosyTag struct {
	all          string
	add          string
	update       string
	item         string
	list         []string
	json         string
	batch        bool
	unique       bool
	customFilter *orderedmap.OrderedMap[string, string]
}

// NewCosyTag creates a new CosyTag from a tag string
func NewCosyTag(tag string) (c CosyTag) {
	if tag == "" {
		return
	}

	c.customFilter = orderedmap.NewOrderedMap[string, string]()

	// split tag by ;
	for group := range strings.SplitSeq(tag, ";") {
		// now the group is like "add:required,max=100"
		// we need to get the right side of :
		directives := strings.Split(group, ":")

		// fixed for cosy:"batch", cosy:"db_unique"
		if len(directives) == 1 {
			directives = append(directives, "")
		}
		if len(directives) < 2 {
			continue
		}

		// now the directives are like
		// ["all", "omitempty"]
		// ["add", "required,max=100"],
		// ["list", "fussy"]
		// ["update", "omitempty"]
		// ["item", "preload"]
		// ["json", "password"]
		// ["list", "fussy[sakura]"]

		switch directives[0] {
		// for "add", "update", "item" directives, we only need the right side
		case "all":
			c.all = directives[1]
		case "add":
			c.add = directives[1]
		case "update":
			c.update = directives[1]
		case "item":
			c.item = directives[1]
		// for list directives, we need to split the right side by ,
		case "list":
			c.list = strings.Split(directives[1], ",")
		case "json":
			c.json = directives[1]
		// for batch directives, we only need the left side
		case "batch":
			c.batch = true
		case "db_unique":
			c.unique = true
		}
	}

	return c
}

// GetAdd returns the add directive
func (c *CosyTag) GetAdd() string {
	if c.all == "" {
		return c.add
	}
	var sb strings.Builder
	if c.add != "" {
		sb.WriteString(c.add)
	}
	if c.add != "" && c.all != "" {
		sb.WriteString(",")
	}
	if c.all != "" {
		sb.WriteString(c.all)
	}
	return sb.String()
}

// GetUpdate returns the update directive
func (c *CosyTag) GetUpdate() string {
	if c.all == "" {
		return c.update
	}
	var sb strings.Builder
	if c.update != "" {
		sb.WriteString(c.update)
	}
	if c.update != "" && c.all != "" {
		sb.WriteString(",")
	}
	if c.all != "" {
		sb.WriteString(c.all)
	}
	return sb.String()
}

// GetItem returns the item directive
func (c *CosyTag) GetItem() string {
	return c.item
}

// GetList returns the list directive
func (c *CosyTag) GetList() []string {
	return c.list
}

// GetJson returns the JSON directive
func (c *CosyTag) GetJson() string {
	return c.json
}

// GetBatch returns the batch directive
func (c *CosyTag) GetBatch() bool {
	return c.batch
}

// GetUnique returns the unique directive
func (c *CosyTag) GetUnique() bool {
	return c.unique
}
