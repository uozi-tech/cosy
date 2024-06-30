# Sandbox

Sandbox 是 Cosy 的优化测试方案，可以有效简化您的测试代码，提高测试效率。

在 Sandbox 实例中，我们会自动完成测试环境的准备工作，例如：
 - 初始化数据库，配置一个随机的表前缀避免多个测试运行时出现数据表冲突
 - 迁移表结构
 - 连接 Redis（如有）
 - 在测试结束后清理测试环境，包括删除表和 Redis Keys <sup>[1]</sup>。

示例代码如下，我们先新建一个示例，然后注册模型，最后在 Run 函数中编写您的测试用例。

```go
package cosy

import (
	"git.uozi.org/uozi/cosy/router"
	"git.uozi.org/uozi/cosy/sandbox"
	"testing"
)

func TestSandbox(t *testing.T) {
	sandbox.NewInstance("app.ini", "pgsql").
		RegisterModels(User{}).
		Run(func(instance *sandbox.Instance) {
			r := router.GetEngine()
			g := r.Group("/")
			Api[User]("users").InitRouter(g)
			
            resp, err = c.Get("/users/1")
            if err != nil {
                t.Error(err)
                return
            }
			var user User
			err = resp.To(&user)
            if err != nil {
                t.Error(err)
                return
            }
            assert.Equal(t, http.StatusOK, resp.StatusCode)
		    assert.Equal(t, uint64(1), user.ID)
		})
}
```

## Sandbox 接口参考

### 新建实例
```go
func NewInstance(configPath, databaseType string) *Instance
```
databaseType 可选值：
  * mysql
  * pgsql
  * sqlite

### 注册模型
```go
func (instance *Instance) RegisterModels(models ...interface{}) *Instance
```

### 运行测试用例
```go
func (instance *Instance) Run(f func(instance *Instance))
```

### 获取测试请求客户端
```go
func (instance *Instance) GetClient() *Client
```

## Client 接口参考
### 添加请求 Header
```go
func (c *Client) AddHeader(key, value string)
```

### 发送 GET 请求
```go
func (c *Client) Get(url string) (*Response, error)
```

### 发送 POST 请求
```go
func (c *Client) Post(url string, body any) (*Response, error)
```

### 发送 PUT 请求
```go
func (c *Client) Put(url string, body any) (*Response, error)
```

### 发送 DELETE 请求
```go
func (c *Client) Delete(url string) (*Response, error)
```

### 发送 PATCH 请求
```go
func (c *Client) Patch(url string, body any) (*Response, error)
```

### 发送 OPTIONS 请求
```go
func (c *Client) Options(url string) (*Response, error)
```

## 响应接口参考
```go
type Response struct {
	StatusCode int
	body       []byte
}
```

### 将响应体解码为 JSON
```go
func (r *Response) To(dest any) error
```




***
[1] 清理环境基于 Run 函数的 defer，如果出现 Fatal 等不可 recover 的错误，可能会导致清理环境失败。