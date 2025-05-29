# Cron 定时任务
Cosy 使用 [go-co-op/gocron](https://github.com/go-co-op/gocron) 实现定时任务调度。

## 注册定时任务
```go
cosy.RegisterCronJob("task-name", func(s gocron.Scheduler) {
    // 定义定时任务
    s.NewJob(
        gocron.DurationJob(time.Minute*5), // 每5分钟执行一次
        gocron.NewTask(func() {
            // 任务逻辑
            logger.Info("定时任务执行")
        }),
    )
})
```

## 基本用法

### 按时间间隔执行

```go
cosy.RegisterCronJob("interval-task", func(s gocron.Scheduler) {
    // 每30秒执行一次
    s.NewJob(
        gocron.DurationJob(time.Second*30),
        gocron.NewTask(func() {
            logger.Info("每30秒执行的任务")
        }),
    )

    // 每小时执行一次
    s.NewJob(
        gocron.DurationJob(time.Hour),
        gocron.NewTask(func() {
            logger.Info("每小时执行的任务")
        }),
    )
})
```

### 按Cron表达式执行

```go
cosy.RegisterCronJob("cron-task", func(s gocron.Scheduler) {
    // 每天凌晨2点执行
    s.NewJob(
        gocron.CronJob("0 2 * * *", false),
        gocron.NewTask(func() {
            logger.Info("每天凌晨2点执行的任务")
        }),
    )

    // 每周一上午9点执行
    s.NewJob(
        gocron.CronJob("0 9 * * 1", false),
        gocron.NewTask(func() {
            logger.Info("每周一上午9点执行的任务")
        }),
    )
})
```

### 带参数的任务

```go
cosy.RegisterCronJob("task-with-params", func(s gocron.Scheduler) {
    s.NewJob(
        gocron.DurationJob(time.Minute*10),
        gocron.NewTask(func(message string, count int) {
            logger.Info("带参数的任务", "message", message, "count", count)
        }, "Hello World", 42),
    )
})
```

## 高级功能

### 任务选项配置

```go
cosy.RegisterCronJob("advanced-task", func(s gocron.Scheduler) {
    job, err := s.NewJob(
        gocron.DurationJob(time.Minute*5),
        gocron.NewTask(func() {
            logger.Info("高级任务执行")
        }),
        gocron.WithName("my-advanced-task"),           // 设置任务名称
        gocron.WithTags("important", "daily"),         // 设置标签
        gocron.WithStartAt(gocron.WithStartImmediately()), // 立即开始
        gocron.WithEventListeners(
            gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
                logger.Info("任务开始执行", "jobID", jobID, "jobName", jobName)
            }),
            gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
                logger.Info("任务执行完成", "jobID", jobID, "jobName", jobName)
            }),
        ),
    )
    if err != nil {
        logger.Error("创建任务失败", "error", err)
    }
})
```

### 错误处理

```go
cosy.RegisterCronJob("error-handling-task", func(s gocron.Scheduler) {
    s.NewJob(
        gocron.DurationJob(time.Minute),
        gocron.NewTask(func() error {
            // 可能出错的任务逻辑
            if someCondition {
                return errors.New("任务执行失败")
            }
            logger.Info("任务执行成功")
            return nil
        }),
        gocron.WithEventListeners(
            gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
                logger.Error("任务执行出错", "jobID", jobID, "jobName", jobName, "error", err)
            }),
        ),
    )
})
```

## 完整示例

```go
package main

import (
    "time"
    "github.com/go-co-op/gocron/v2"
    "github.com/uozi-tech/cosy"
    "github.com/uozi-tech/cosy/logger"
)

func main() {
    // 注册数据清理任务
    cosy.RegisterCronJob("data-cleanup", func(s gocron.Scheduler) {
        s.NewJob(
            gocron.CronJob("0 3 * * *", false), // 每天凌晨3点执行
            gocron.NewTask(func() {
                logger.Info("开始清理过期数据")
                // 清理逻辑
                cleanupExpiredData()
                logger.Info("数据清理完成")
            }),
            gocron.WithName("daily-cleanup"),
        )
    })

    // 注册健康检查任务
    cosy.RegisterCronJob("health-check", func(s gocron.Scheduler) {
        s.NewJob(
            gocron.DurationJob(time.Minute*5), // 每5分钟执行一次
            gocron.NewTask(func() {
                // 健康检查逻辑
                if !checkSystemHealth() {
                    logger.Error("系统健康检查失败")
                    // 发送告警
                    sendAlert("系统健康检查失败")
                }
            }),
            gocron.WithName("health-monitor"),
        )
    })

    // 启动应用
    cosy.Boot("app.ini")
}

func cleanupExpiredData() {
    // 实现数据清理逻辑
}

func checkSystemHealth() bool {
    // 实现健康检查逻辑
    return true
}

func sendAlert(message string) {
    // 实现告警发送逻辑
}
```

## 常用Cron表达式

| 表达式 | 说明 |
|--------|------|
| `0 0 * * *` | 每天午夜执行 |
| `0 */6 * * *` | 每6小时执行一次 |
| `0 9 * * 1-5` | 工作日上午9点执行 |
| `0 0 1 * *` | 每月1号午夜执行 |
| `0 0 * * 0` | 每周日午夜执行 |

## 注意事项

1. 定时任务在应用启动时自动开始执行
2. 任务执行是异步的，不会阻塞主程序
3. 建议为重要任务添加错误处理和日志记录
4. 长时间运行的任务应该考虑超时控制
5. 在生产环境中，建议使用分布式锁避免多实例重复执行
