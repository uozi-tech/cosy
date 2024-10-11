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

<div style="display: flex;justify-content: center;">
    <img src="/assets/item.png" alt="list" style="max-width: 500px;width: 95%"/>
</div>

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

## 排序和分页
Query 请求参数说明
- sort_by: 排序字段
- order: desc 倒序，asc 顺序
- page: 当前页数
- page_size: 每页数量

::: tip 提示
为了避免数据库注入，只有 Struct 定义了的字段才可以排序，如果你使用了 SQL View 扩展了字段，
可以调用 `AddColWhiteList(cols ...string)` 方法，将这些字段加入白名单。
:::

## 其他方法

以下方法的使用与获取**单个记录**的方式相同

- SetTable(table string)
- SetTransformer(fx func(user *model.User) any)
- SetScan(fx func(tx *gorm.DB) any)
- GormScope(fx func(tx *gorm.DB) *gorm.DB)

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
    "last_page": 1
  }
}
```