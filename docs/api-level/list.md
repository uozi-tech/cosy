# 列表

```go
func GetList() {
   core := cosy.Core[model.User](c).
   SetFussy("name", "phone", "email").
   SetIn("status")

   core.PagingList()
}
```

## 生命周期

1. **BeforeExecute**
2. 执行获取操作
3. **Executed**
4. 返回响应

```mermaid
flowchart TD
  A[请求到达] --> P[Prepare: getListHook + 预加载与关联 + prepareHook]
  P --> BE[BeforeExecute Hook]
  BE --> RES[结果集: Model 与 T]
  RES --> TRASH{是否使用 trash 参数?}
  TRASH -- 是 --> UNSC[Unscoped 并过滤已删除]
  TRASH -- 否 --> KEEP[正常查询]
  UNSC --> PAG[分页查询]
  KEEP --> PAG
  PAG --> COUNT[统计总数 移除排序与限制]
  COUNT --> EX[Executed Hook]
  EX --> RESP[200 OK 返回数据和分页]
```

## 筛选方法

::: tip 提示
筛选方法可以被多次调用，本质上执行的是 slice 的 `append` 方法。
:::

1. SetFussy(keys ...string)
    - 设置模糊搜索, 使用 LIKE %...% 作为查询条件。
2. SetEqual(keys ...string)
    - 设置等于查询, 使用 = 作为查询条件。
3. SetIn(keys ...string)
    - 设置 IN 查询, 使用 IN 作为查询条件。
4. SetOrFussy(keys ...string)
    - 设置模糊搜索的 OR 查询, 使用 LIKE %...% 或者其他条件。
5. SetOrEqual(keys ...string)
    - 设置等于查询的 OR 查询, 使用 = 或者其他条件。
6. SetOrIn(keys ...string)
    - 设置 IN 查询的 OR 查询, 使用 IN 或者其他条件。
7. SetSearchFussyKeys(keys ...string)
    - 设置多个字段的模糊搜索，使用子查询 OR 连接。

## Query 参数与数据库列的自动映射

如果模型使用的是 `json:"camelCase"` + `gorm:"column:snake_case"`，那么列表查询时可以直接使用 camelCase 参数，Cosy 会自动映射到数据库列。

```go
type Project struct {
   Model
   EnvironmentID string    `json:"environmentId" cosy:"list:eq" gorm:"column:environment_id;index"`
   CreatedByID   string    `json:"createdById" cosy:"list:eq" gorm:"column:created_by_id;index"`
   CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at"`
}
```

请求示例：

```text
GET /projects?environmentId=prod&createdById=u_123&sort_by=createdAt&order=desc
```

实际查询会使用数据库列：

- `environmentId` -> `environment_id`
- `createdById` -> `created_by_id`
- `createdAt` -> `created_at`

:::: tip 提示
`SetEqual("environmentId")`、`SetIn("createdById")`、`SetBetween("createdAt")` 等方法，都会优先按 query key 读取请求参数，再按模型中的 GORM 列名拼接 SQL。
::::

## 排序和分页
Query 请求参数说明
- sort_by: 排序字段
- order: desc 倒序，asc 顺序
- page: 当前页数
- page_size: 每页数量

:::: tip 提示
如需同时查看已软删除的数据，可在查询参数中加入 `trash=true`。
::::

如果你的 API 使用 camelCase JSON，可以直接传 `sort_by=createdAt` 这类参数，Cosy 会自动映射为对应的数据库列 `created_at`。

## 非分页列表
当数据量较小或需要一次性返回全部数据时，可使用 `List()`：

```go
func GetAllUsers(c *gin.Context) {
   cosy.Core[model.User](c).
      SetFussy("name", "phone").
      SetIn("status").
      List()
}
```

返回为纯数组，未包含分页信息。

## 空分页响应
当需要返回一个空的分页结构（例如首次加载或无数据时），可使用 `EmptyPagingList()`：

```go
func GetEmpty(c *gin.Context) {
   cosy.Core[model.User](c).EmptyPagingList()
}
```

返回示例：

```json
{
  "data": [],
  "pagination": {
    "per_page": 10
  }
}
```

如果你希望分页字段也输出 camelCase，可以在构建时使用 `camelcase_json` build tag：

```shell
go build -tags=camelcase_json ./...
```

开启后分页字段会变为：

```json
{
  "data": [],
  "pagination": {
    "perPage": 10
  }
}
```

## 标准选择器初始化
当请求中包含 `id[]` 参数时，列表会优先按这些 ID 过滤，常用于「标准选择器」初始化：

```
GET /users?page=1&id[]=1&id[]=2&id[]=3
```

::: tip 提示
为了避免数据库注入，只有 Struct 定义了的字段才可以排序，如果你使用了 SQL View 扩展了字段，
可以调用 `AddColWhiteList(cols ...string)` 方法，将这些字段加入白名单。
:::

如果您需要使用自己的逻辑进行排序，请使用 `func WithoutSortOrder() *Ctx[T]` 方法禁用默认的排序逻辑。

## 其他方法

以下方法的使用与获取**单个记录**的方式相同

- SetTable(table string)
- SetTransformer(fx func(user *model.User) any)
- SetScan(fx func(tx *gorm.DB) any)
- GormScope(fx func(tx *gorm.DB) *gorm.DB)
- SetResponseBuilder(func(ctx *cosy.Ctx[model.User]))

## 自定义响应构建
当你需要对列表结果做整体后处理时，可以使用 `SetResponseBuilder` 接管最终响应。

```go
func GetUsers(c *gin.Context) {
   cosy.Core[model.User](c).
      SetFussy("name", "phone").
      SetResponseBuilder(func(ctx *cosy.Ctx[model.User]) {
         // 原始默认响应（PagingList 场景包含 pagination）
         origin := ctx.GetDefaultResponseData()
         listData, _ := origin.(model.DataList)

         // 在这里仅改 data，不丢分页
         ctx.JSON(http.StatusOK, model.DataList{
            Data:       listData.Data,
            Pagination: listData.Pagination,
         })
      }).
      PagingList()
}
```

::: tip 提示
`SetResponseBuilder` 不会改变查询逻辑，只会覆盖最终输出。若未设置，仍使用 Cosy 默认响应格式。你也可以通过 `ctx.GetDefaultResponseData()` 获取原始默认响应，避免分页等信息丢失。
:::

## 响应示例

```json
{
  "data": [
    {
      "id": 1,
      "name": "Jacky",
      "email": "me@jackyu.cn",
      "phone": "123456789",
      "avatar": "avatar.jpg",
      "last_active": "2024-01-01T00:00:00Z",
      "power": 1,
      "status": 1,
      "group_id": 1,
      "group": {
        "id": 1,
        "name": "Admin"
      },
      "group_name": "Admin"
    }
  ],
  "pagination": {
    "total": 1,
    "per_page": 10,
    "current_page": 1,
    "total_pages": 1
  }
}
```
