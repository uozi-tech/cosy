# 模型定义

入门指南将以一个简单的 User CURD 为例，首先我们为他定义一个模型：

一般我们会将模型文件统一放在 `model` 目录下

```go
package model

import (
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        int             `gorm:"primary_key" json:"id"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type Group struct {
	Model
	Name string `json:"name"`
}

type User struct {
	Model
	Name       string     `json:"name"`
	Password   string     `json:"-"` // hide password
	Email      string     `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" gorm:"index"`
	Avatar     string     `json:"avatar"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" gorm:"default:1"`
	Status     int        `json:"status" gorm:"default:1"`
	GroupID    int        `json:"group_id"`
	Group      *Group     `json:"group"`
}
```