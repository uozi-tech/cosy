# 热更新
你可以直接修改当前项目和 Cosy 包内的 settings 中的结构体字段的值。

如果需要将更新的设置持久化，可以使用 `func Save() (err error)`。

## 示例
```go
package main

import (
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    err := settings.Save()
    if err != nil {
        logger.Error(err)
        return 
    }
}
```
