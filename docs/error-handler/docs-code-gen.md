# 文档和代码生成

## 使用方法

```bash
go run cmd/errdocs/generate.go -project <项目路径> -type <类型> -output <输出目录> [-wrapper <包装函数>] [-trailing-comma <是否添加逗号>] [-ignore-dirs <忽略目录>]
```

| 参数 | 说明 | 是否必填 | 默认值 |
| --- | --- | --- | --- |
| -project | 项目根目录路径 | 是 | - |
| -type | 生成文件类型：`md`/`ts`/`js` | 是 | - |
| -output | 输出目录 | 是 | - |
| -wrapper | 错误信息包装函数 | 否 | `$gettext` |
| -trailing-comma | 是否在最后一项添加逗号 | 否 | `true` |
| -ignore-dirs | 要忽略的目录（逗号分隔） | 否 | 空 |

示例：
```bash
# 生成 Markdown 文档
go run cmd/errdocs/generate.go -project ./project -type md -output ./docs

# 生成 TypeScript 错误定义，使用自定义 wrapper 并忽略 vendor/test 目录
go run cmd/errdocs/generate.go -project ./project -type ts -output ./src/errors -wrapper t -ignore-dirs vendor,test

# 生成 JavaScript 错误定义，不使用末尾逗号并忽略 node_modules
go run cmd/errdocs/generate.go -project ./project -type js -output ./dist -wrapper '$gettext' -trailing-comma=false -ignore-dirs node_modules
```

从 `v1.14.2` 开始，您可以在项目中的 `cmd/errdef/generate.go` 中使用 `error.Generate()` 来创建生成器，以实现在本地调用文档和代码生成器。

## 生成的文件示例

### Markdown 文档
```markdown
# auth

| Error Code | Error Message |
| --- | --- |
| 4031 | Token is empty |
| 4032 | Token convert to claims failed |
| -4033 | JWT expired |
```

### TypeScript 错误定义
```ts
export default {
  4031: () => $gettext('Token is empty'),
  4032: () => $gettext('Token convert to claims failed'),
  '-4033': () => $gettext('JWT expired'),
}
```

### JavaScript 错误定义
```js
module.exports = {
  4031: () => $gettext('Token is empty'),
  4032: () => $gettext('Token convert to claims failed'),
  '-4033': () => $gettext('JWT expired'),
}
```

注意：
1. 错误信息的首字母会自动转换为大写
2. 负数错误码在 TypeScript 和 JavaScript 中会使用字符串形式
3. 包含引号的错误信息会自动进行转义处理
4. 默认会在最后一项添加逗号，可以通过参数禁用

