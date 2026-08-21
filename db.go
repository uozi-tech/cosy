package cosy

import (
	"context"

	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

// UseDB return the ptr of gorm.DB.
//
// ctx may be a *gin.Context: it is detached via RequestContext before being
// attached to the returned *gorm.DB, so it is safe to call UseDB(c) inside a
// handler.
func UseDB(ctx context.Context) *gorm.DB {
	return model.UseDB(ctx)
}

// RequestContext returns a context that is safe to hand to GORM (or any other
// API that may keep goroutines alive after the handler returns) in place of a
// *gin.Context. Use it when attaching a context to a *gorm.DB you manage
// yourself: db.WithContext(cosy.RequestContext(c)).
//
// Never pass a *gin.Context straight to GORM: gin pools and reuses Context
// objects across requests, and database/sql goroutines (Rows.awaitDone) may
// still call ctx.Done()/ctx.Err() on it after the request has finished, racing
// with the next request that reuses the same Context. See model.RequestContext.
func RequestContext(ctx context.Context) context.Context {
	return model.RequestContext(ctx)
}

// RegisterModels register models.
func RegisterModels(models ...any) {
	model.RegisterModels(models...)
}

// InitDB init db.
func InitDB(dialect gorm.Dialector) *gorm.DB {
	return model.Init(dialect)
}
