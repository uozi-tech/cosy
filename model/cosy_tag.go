package model

import "strings"

type CosyTag struct {
	all    string
	add    string
	update string
	item   string
	list   []string
	json   string
}

func NewCosyTag(tag string) (c CosyTag) {
	if tag == "" {
		return
	}

	// split tag by ;
	groups := strings.Split(tag, ";")
	for _, group := range groups {
		// now the group is like "add:required,max=100"
		// we need to get the right side of :
		directives := strings.Split(group, ":")
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

		switch directives[0] {
		// for add, update, item directives, we only need the right side
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
		}
	}

	return c
}

func (c *CosyTag) GetAdd() string {
	if c.all == "" {
		return c.add
	}
	var sb strings.Builder
	if c.add != "" {
		sb.WriteString(c.add)
	}
	if c.add != "" {
		sb.WriteString(",")
		sb.WriteString(c.all)
	}
	return sb.String()
}

func (c *CosyTag) GetUpdate() string {
	if c.add == "" {
		return c.update
	}
	var sb strings.Builder
	if c.update != "" {
		sb.WriteString(c.update)
	}
	if c.all != "" {
		sb.WriteString(",")
		sb.WriteString(c.all)
	}
	return sb.String()
}

func (c *CosyTag) GetItem() string {
	return c.item
}

func (c *CosyTag) GetList() []string {
	return c.list
}

func (c *CosyTag) GetJson() string {
	return c.json
}
