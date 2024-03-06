package model

import "gorm.io/gen"

type Method interface {
	// FirstByID Where("id=@id")
	FirstByID(id int) (*gen.T, error)
	// DeleteByID update @@table set deleted_at=strftime('%Y-%m-%d %H:%M:%S','now') where id=@id
	DeleteByID(id int) error
	// RecoverByID update @@table set deleted_at=NULL where id=@id
	RecoverByID(id int) error
}
