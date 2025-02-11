# Cron 定时任务
Cosy 使用 [go-co-op/gocron](https://github.com/go-co-op/gocron) 实现定时任务调度。

## 注册定时任务
```go
cosy.RegisterCronJob("task-name", func(s gocron.Scheduler) {
    // your code here
})
```
