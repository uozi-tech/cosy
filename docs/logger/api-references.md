# API 参考

## 初始化
```go
func Init(mode string)
```

初始化日志系统，其中`mode`参数可以是：
- `gin.DebugMode`: 调试模式，输出所有日志级别
- `gin.ReleaseMode`: 发布模式，仅输出Info级别及以上日志

## 获取实例
```go
func GetLogger() *zap.SugaredLogger
```

该方法通常用于，提供给用户一个方案，使得用户可以自由一个新的 logger 函数。

通常在封装函数内调用 logger 时，输出的调用栈是封装内 logger 的行号

```go
logger.GetLogger().WithOptions(zap.AddCallerSkip(1)).Errorln(err)
```

使用 `WithOptions(zap.AddCallerSkip(1))` 可以使得输出的调用栈是调用该封装函数的文件和行号。

## 同步缓冲区
```go
func Sync()
```

同步并刷新所有缓冲的日志条目。在程序退出前调用可确保所有日志被写入。

## Debug
```go
func Debug(args ...interface{})
```

## Debugf
```go
func Debugf(format string, args ...interface{})
```

## Info
```go
func Info(args ...interface{})
```

## Infof
```go
func Infof(format string, args ...interface{})
```

## Warn
```go
func Warn(args ...interface{})
```

## Warnf
```go
func Warnf(format string, args ...interface{})
```

## Error
```go
func Error(args ...interface{})
```

## Errorf
```go
func Errorf(format string, args ...interface{})
```

## Fatal
```go
func Fatal(args ...interface{})
```

## Fatalf
```go
func Fatalf(format string, args ...interface{})
```

## Panic
```go
func Panic(args ...interface{})
```

## Panicf
```go
func Panicf(format string, args ...interface{})
```

## DPanic
```go
func DPanic(args ...interface{})
```

## DPanicf
```go
func DPanicf(format string, args ...interface{})
```
