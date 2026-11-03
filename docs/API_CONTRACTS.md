# API 契约与变更维护

本文只描述稳定的跨端规则；**不再手工复制逐接口字段表**。接口数、路径、请求/响应字段和权限应从同一提交的 Go 路由与生成产物读取，避免静态文档落后于实际代码。

## 权威来源

| 内容 | 来源 |
| --- | --- |
| HTTP 方法、路径、权限边界 | `智慧养老手表管理系统/backend/internal/accesspolicy/policy.go` 与各 `handlers/handler_*.go` |
| 请求/响应字段清单 | `智慧养老手表管理系统/backend/internal/apidoc/contracts.json` |
| OpenAPI 生成模板和服务端动态输出 | `智慧养老手表管理系统/backend/internal/apidoc/openapi-template.json`、`GET /openapi.json` |
| 浏览器 HTTP 响应与错误处理 | `智慧养老手表管理系统/frontend/src/api/httpContract.ts` |
| 手表设备通信 | [手表工程说明](../SuiningWatch/README.md)及生产 ArkTS `WatchApi.ets` |

`GET /openapi.json` 合并路由、共享 schema 和字段清单。源码变更后执行：

```sh
python3 scripts/generate-openapi.py
python3 scripts/generate-openapi.py --check
```

第一条更新生成产物，第二条检查一致性。类型、可空性及状态码变化直接更新 schema 与回归测试。

## 路由与客户端约定

- 业务使用 `/api/v1/`。管理页面在 `/operations`，家属页面在 `/family-portal`，登录和家属注册入口分别为 `/sign-in`、`/join-family`。不提供旧版无版本号的接口别名。
- 浏览器依赖同源会话 `SUINING_SESSION` 与写请求 CSRF 校验；用户身份从服务端会话获得，不接受客户端提交的姓名、手机号替代权限校验。
- 设备接口使用设备号和密钥认证，并按接口要求校验绑定版本、事件 ID 与重试语义。不同的网络异常、认证失败及绑定冲突不可混为“已接收”。
- JSON 成功通常返回 HTTP 200 和根级字符串 `returnCode: "1111"`，业务字段直接位于根对象。业务错误使用相应 HTTP 4xx/5xx，并提供 `code`、`message`；内部错误带追踪 `reference`。不同业务端点的附加字段以 schema 与测试为准。
- 列表分页请求一般为 `page`、`pageSize`，响应为 `total`、`page`、`pageSize` 和相应业务列表。空集合应返回数组而非对象；允许空值的字段以 schema 为准。
- `POST /api/v1/image/saveImage` 使用 multipart `file`；`GET /api/v1/image/download?name=...` 成功返回图片二进制，不套 JSON 结果包。图片访问受鉴权与存储完整性检查保护。
- 健康时间使用带偏移的 RFC3339，原始采样、位置采样与事件接收时间须区分；不把历史位置冒充本次 SOS 现场坐标。

## HTTP 状态分类

| 状态 | 常见含义 |
| --- | --- |
| 400 / 413 / 415 | 字段无效、请求过大、媒体类型不受支持 |
| 401 / 403 | 未认证或无权访问、设备停用、CSRF 拒绝 |
| 404 / 405 | 资源/路由不存在、HTTP 方法不允许；405 应包含 `Allow` |
| 409 | 幂等键、绑定代次、资源修订或业务状态冲突 |
| 429 | 触发限流 |
| 500 / 503 | 服务内部错误、数据库或所需基础设施不可用 |

上表并不强制每条路径返回所有状态，也不替代端点自身的 `code`。敏感错误不得回显 SQL、密钥或隐私数据。

## 修改接口时的检查

1. 修改 Go 路由、请求 DTO、响应或权限时，同步更新 `accesspolicy`、`contracts.json` 和必要的 OpenAPI schema。
2. 同步前端调用、穿戴协议、状态码字典和回归测试。
3. 校验生成产物、后端测试、前端 typecheck/测试和实际受影响的协议集成路径；对数据库、签名、幂等及事件归属变更另外执行专门的隔离环境验证。
4. 记录变更 SHA、实际验证项与测试结果。
