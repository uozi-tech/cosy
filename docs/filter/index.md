# 筛选器

从 `v1.13.0` 版本开始，我们引入自定义列表筛选器的能力。

## 实现筛选器
```go
package crypto

import (
  "github.com/gin-gonic/gin"
  "github.com/uozi-tech/cosy/model"
  "gorm.io/gorm"
)

type MyFilter struct{}

func (MyFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
  queryValue := c.Query(key)
  if queryValue == "" {
    return db
  }
  // ${FieldName}_like: name_like, phone_like
  myColumn := key + "_like"
  return myGormSearch(myColumn, queryValue)(db)
}
```

## 注册筛选器
```go
func init() {
    filter.RegisterFilter("fussy[my]", MyFilter{})
}
```

## 使用筛选器
定义模型时使用

```go
package model

type MyModel struct {
  Name string `json:"name" cosy:"list:fussy[my]"`
}
```
