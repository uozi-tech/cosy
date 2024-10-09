# Logger

封装 go.uber.org/zap

对于接口级简化的项目，需要手动初始化 Logger，如下：

```go

import (
    "github.com/uozi-tech/cosy/logger"
)

func main() {
    // ...
    logger.Init()
	defer logger.Sync()
    // ...
}
```

对于项目级简化的项目，无需手动初始化。
