# 文档和代码生成

## 使用方法

```bash
go run cmd/errdocs/generate.go <project folder path> <type: md|ts|js> <output dir> [wrapper] [trailing_comma]
```

| 参数 | 说明 | 是否必填 | 默认值 |
| --- | --- | --- | --- |
| project folder path | 项目根目录路径 | 是 | - |
| type | 生成文件类型：`md`/`ts`/`js` | 是 | - |
| output dir | 输出目录 | 是 | - |
| wrapper | 错误信息包装函数 | 否 | `$gettext` |
| trailing_comma | 是否在最后一项添加逗号 | 否 | `true` |

示例：
```bash
# 生成 Markdown 文档
go run cmd/errdocs/generate.go ./project md ./docs

# 生成 TypeScript 错误定义，使用自定义 wrapper
go run cmd/errdocs/generate.go ./project ts ./src/errors t

# 生成 JavaScript 错误定义，不使用末尾逗号
go run cmd/errdocs/generate.go ./project js ./dist $gettext false
```

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

