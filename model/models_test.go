package model

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestEmbed struct {
	Avatar string `json:"avatar" cosy:"all:omitempty"`
}

type TestEmbedPtr struct {
	AuditStatus int `json:"audit_status" cosy:"all:omitempty"`
}

type User struct {
	Model
	Name     string `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password string `json:"-" cosy:"add:required;update:omitempty"` // hide password
	Email    string `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"uniqueIndex"`
	Phone    string `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	TestEmbed
	*TestEmbedPtr
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" cosy:"add:required;update:omitempty;list:in" gorm:"default:1;index"`
	Status     int        `json:"status" cosy:"add:required;update:omitempty;list:in" gorm:"default:1;index"`
	Group      string     `json:"group" cosy:"add:required;update:omitempty;list:in" gorm:"index"`
}

type Product struct {
	Model
	Name        string          `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Description string          `json:"description" cosy:"add:required;update:omitempty;list:fussy"`
	Price       decimal.Decimal `json:"price" cosy:"add:required;update:omitempty;list:fussy"`
	Status      string          `json:"status" cosy:"add:required;update:omitempty;list:in"`
	UserID      int             `json:"user_id" gorm:"index"`
	User        *User           `json:"user,omitempty" cosy:"item:preload"`
}

func TestResolvedModels(t *testing.T) {
	RegisterModels(User{}, Product{})

	ResolvedModels()

	expectedModel := map[string]ResolvedModel{
		"User": {
			Name: "User",
			OrderedFields: []*resolvedModelField{
				{
					Name:    "ID",
					Type:    "int",
					JsonTag: "id",
					CosyTag: CosyTag{},
				},
				{
					Name:    "CreatedAt",
					Type:    "time.Time",
					JsonTag: "created_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "UpdatedAt",
					Type:    "time.Time",
					JsonTag: "updated_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "DeletedAt",
					Type:    "*gorm.DeletedAt",
					JsonTag: "deleted_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "Name",
					Type:    "string",
					JsonTag: "name",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Password",
					Type:    "string",
					JsonTag: "-",
					CosyTag: NewCosyTag("add:required;update:omitempty"),
				},
				{
					Name:    "Email",
					Type:    "string",
					JsonTag: "email",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Phone",
					Type:    "string",
					JsonTag: "phone",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Avatar",
					Type:    "string",
					JsonTag: "avatar",
					CosyTag: NewCosyTag("all:omitempty"),
				},
				{
					Name:    "AuditStatus",
					Type:    "int",
					JsonTag: "audit_status",
					CosyTag: NewCosyTag("all:omitempty"),
				},
				{
					Name:    "LastActive",
					Type:    "*time.Time",
					JsonTag: "last_active",
					CosyTag: CosyTag{},
				},
				{
					Name:    "Power",
					Type:    "int",
					JsonTag: "power",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:in"),
				},
				{
					Name:    "Status",
					Type:    "int",
					JsonTag: "status",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:in"),
				},
				{
					Name:    "Group",
					Type:    "string",
					JsonTag: "group",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:in"),
				},
			},
		},
		"Product": {
			Name: "Product",
			OrderedFields: []*resolvedModelField{
				{
					Name:    "ID",
					Type:    "int",
					JsonTag: "id",
					CosyTag: CosyTag{},
				},
				{
					Name:    "CreatedAt",
					Type:    "time.Time",
					JsonTag: "created_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "UpdatedAt",
					Type:    "time.Time",
					JsonTag: "updated_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "DeletedAt",
					Type:    "*gorm.DeletedAt",
					JsonTag: "deleted_at",
					CosyTag: CosyTag{},
				},
				{
					Name:    "Name",
					Type:    "string",
					JsonTag: "name",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Description",
					Type:    "string",
					JsonTag: "description",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Price",
					Type:    "decimal.Decimal",
					JsonTag: "price",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:fussy"),
				},
				{
					Name:    "Status",
					Type:    "string",
					JsonTag: "status",
					CosyTag: NewCosyTag("add:required;update:omitempty;list:in"),
				},
				{
					Name:    "UserID",
					Type:    "int",
					JsonTag: "user_id",
					CosyTag: CosyTag{},
				},
				{
					Name:    "User",
					Type:    "*model.User",
					JsonTag: "user",
					CosyTag: NewCosyTag("item:preload"),
				},
			},
		},
	}

	assert := assert.New(t)
	for name, model := range resolvedModelMap {
		assert.Equal(expectedModel[name].Name, model.Name)
		for k, field := range model.OrderedFields {
			assert.Equal(expectedModel[name].OrderedFields[k].Name, field.Name)
			assert.Equal(expectedModel[name].OrderedFields[k].Type, field.Type)
			assert.Equal(expectedModel[name].OrderedFields[k].JsonTag, field.JsonTag)
			assert.Equal(expectedModel[name].OrderedFields[k].CosyTag, field.CosyTag)
		}
	}
}

func TestGetResolvedModel(t *testing.T) {
	RegisterModels(User{}, Product{})

	ResolvedModels()

	assert := assert.New(t)
	user := GetResolvedModel[User]()
	product := GetResolvedModel[Product]()

	assert.Equal("User", user.Name)
	assert.Equal("Product", product.Name)
}
