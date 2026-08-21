package model

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/internal/structcodec"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db            *gorm.DB
	dialectName   string
	beforeMigrate []func(*gorm.DB) error
)

// DialectName returns the name of the database dialect in use (e.g. "postgres",
// "mysql", "sqlite"). It is cached once during Init so callers can branch on
// dialect without paying a per-query lookup.
func DialectName() string {
	return dialectName
}

// BeforeMigrate is a function that will register a function to be executed before db migration
func BeforeMigrate(f func(*gorm.DB) error) {
	beforeMigrate = append(beforeMigrate, f)
}

// logMode return the log mode based on the server run mode
func logMode() gormlogger.Interface {
	switch settings.ServerSettings.RunMode {
	case gin.ReleaseMode:
		return logger.DefaultGormLogger.LogMode(gormlogger.Warn)
	default:
		fallthrough
	case gin.DebugMode:
		return logger.DefaultGormLogger.LogMode(gormlogger.Info)
	}
}

// RequestContext returns a context that is safe to hand to GORM (and, through
// it, database/sql).
//
// A *gin.Context must never be used as the context of a database operation.
// gin pools and reuses Context objects across requests, while database/sql
// keeps goroutines alive that outlive the handler: every query executed inside
// a transaction spawns Rows.awaitDone, which calls ctx.Done()/ctx.Err() after
// the rows are closed. On a *gin.Context those calls read c.Request and
// c.engine, racing with Engine.ServeHTTP writing the same fields for the next
// request that picked the pooled Context.
//
// When ctx is a *gin.Context the result is derived from c.Request.Context(),
// so request-scoped values (session logger, pprof labels, ...) stay visible to
// GORM callbacks and the GORM logger. Cancellation mirrors what the
// *gin.Context itself exposed: with gin's ContextWithFallback disabled (the
// default) a *gin.Context never cancels, so the request context is wrapped in
// context.WithoutCancel; with ContextWithFallback enabled the request context
// is returned as is, keeping client-disconnect cancellation.
//
// Values stored with c.Set are exposed through Value as a snapshot taken when
// RequestContext is called, so GORM hooks reading tx.Statement.Context.Value
// keep working.
//
// When a *gin.Context is buried deeper in the chain (e.g. context.WithValue(c,
// ...)) the outer layers cannot be re-parented, so the whole chain is wrapped in
// context.WithoutCancel: values stay reachable and a caller-supplied deadline
// is re-applied, but Done/Err never touch the pooled *gin.Context.
//
// Any other context is returned unchanged.
func RequestContext(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}

	c, ok := ctx.Value(gin.ContextKey).(*gin.Context)
	if !ok || c == nil {
		return ctx
	}

	if ctx != context.Context(c) {
		detached := context.WithoutCancel(ctx)
		if deadline, ok := ctx.Deadline(); ok {
			// A caller-supplied deadline (context.WithTimeout(c, ...)) must
			// keep bounding the query; only the gin-owned cancellation is cut.
			bounded, cancel := context.WithDeadline(detached, deadline)
			context.AfterFunc(bounded, cancel)
			return bounded
		}
		return detached
	}

	if c.Request == nil {
		return context.Background()
	}

	reqCtx := c.Request.Context()
	if c.Done() == nil {
		// ContextWithFallback is disabled (the default): a *gin.Context never
		// cancels, mirror that instead of inheriting the request cancellation.
		reqCtx = context.WithoutCancel(reqCtx)
	}
	return withGinKeys(c, reqCtx)
}

// ginKeysContext exposes a snapshot of gin's c.Set values through Value so
// GORM hooks and plugins keep reading request data by string key without
// holding a reference to the pooled *gin.Context.
type ginKeysContext struct {
	context.Context
	keys map[any]any
}

func (ctx ginKeysContext) Value(key any) any {
	if value, ok := ctx.keys[key]; ok {
		return value
	}
	return ctx.Context.Value(key)
}

// withGinKeys layers a copy of c.Keys (taken under gin's lock via Copy) over
// parent. Values set on c after this point are not visible.
func withGinKeys(c *gin.Context, parent context.Context) context.Context {
	keys := c.Copy().Keys
	if len(keys) == 0 {
		return parent
	}
	return ginKeysContext{Context: parent, keys: keys}
}

// UseDB return the global db instance.
//
// ctx may be a *gin.Context; it is detached via RequestContext before being
// attached to the returned *gorm.DB so that no database/sql goroutine can
// observe the pooled *gin.Context after the request has finished.
func UseDB(ctx context.Context) *gorm.DB {
	if db == nil {
		return nil
	}
	return db.WithContext(RequestContext(ctx))
}

// Init initialize the global db instance
func Init(dialect gorm.Dialector) *gorm.DB {
	var err error

	db, err = gorm.Open(dialect, &gorm.Config{
		Logger:                                   logMode(),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: settings.DataBaseSettings.TablePrefix,
		},
	})

	if err != nil {
		logger.Fatal(err)
	}

	dialectName = db.Dialector.Name()

	if len(beforeMigrate) > 0 {
		for _, f := range beforeMigrate {
			err = f(db)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	migrate(db, migrationsBeforeAutoMigrate)

	err = db.AutoMigrate(GenerateAllModel()...)

	if err != nil {
		logger.Fatal(err)
	}

	migrate(db, migrationsAfterAutoMigrate)

	ResolvedModels()
	pretouchModels()

	return db
}

// pretouchModels pre-compiles the structcodec decode plan of every registered
// model so the first request does not pay the compilation latency. It runs
// after all registration (and after any converter registration, which
// discards compiled plans).
func pretouchModels() {
	for _, m := range collection {
		if err := structcodec.Pretouch(m); err != nil {
			logger.Error(err)
		}
	}
}
