# 测试、验证与证据分层

验证记录按 Git SHA、数据库类型、操作系统及工具链归档。

## 变更与最小验证范围

| 变更类型 | 最少验证 | 需要增加的验证 |
| --- | --- | --- |
| 仅文档或链接 | 文件路径、内部链接、代码示例与当前源文件一致 | 若调整了 OpenAPI 示例，运行 `python3 scripts/generate-openapi.py --check` |
| Vue 页面、API 类型和状态 | `npm test`、`npm run typecheck`、`npm run build` | 页面交互变化运行对应 Playwright 夹具，验证取消、错误提示和空状态 |
| Go 业务逻辑 | 相关包测试、`go test ./...`、`go vet ./...` | 事务、设备权限、队列和竞态调整需真实 MySQL / PG 与 `-race` |
| 设备/手表协议 | 模拟器与 ArkTS 转译协议测试 | HAP 原生编译、模拟器运行、真机传感器、绑定代次、断网重试分别验证 |
| 数据库迁移或备份恢复 | 独立测试库升级和回滚/恢复路径 | 正式操作前按数据体量做隔离恢复、校验和、写入暂停和签名密钥核对 |
| Windows 安装或签名 | 原生打包、sidecar 哈希、启动及卸载检查 | WebView2 缺失补装、真实桌面交互、目标系统与证书链 |

文档变更检查路径、链接与示例即可。必需检查以[当前 CI](../.github/workflows/ci.yml) 为准。

## Go 后端

在 `智慧养老手表管理系统/backend` 下执行：

```sh
go test -count=1 ./...
go vet ./...
go build ./...
```

对事务、锁序、异步任务或关键安全边界的改动，在**专用可销毁数据库**配置 `TEST_MYSQL_DSN`（MySQL）及 `TEST_DATABASE_URL`（PG 兼容测试），执行真实库 `-race`。例如：

```sh
go test -race -count=1 -timeout 900s ./...
```

CI 以[reusable-backend.yml](../.github/workflows/reusable-backend.yml) 的参数为准，单独运行 PG 配对回归。仅使用隔离测试库，分别记录通过与跳过的测试。

## 前端和穿戴端

```sh
# 在 智慧养老手表管理系统/frontend
npm ci
npm test
npm run typecheck
npm run build

# 在仓库根目录，校验接口规范与生成文件
python3 scripts/generate-openapi.py --check

# 在 SuiningWatch
npm ci
npm test
```

Playwright 在根目录 `qa/` 使用自己的 npm 锁文件：`npm --prefix qa ci`、`npm --prefix qa test`。模拟 API 的页面通过不等价于真实后端。手表 Node 测试使用 ArkTS 适配器，不等价于 DevEco SDK 的真实编译，更不能等价于 WATCH 真机权限、圆屏交互和求救验证。

## 桌面与发布

Windows 构建与签名见 [BUILD](BUILD.md)。检查结果按来源 SHA、构建版本及哈希归档；外部通知和人员响应另记。

## 数据、安全与恢复

- 自动化测试和故障注入只在独立克隆、临时库、测试对象存储及假网关运行，不得覆盖运行中的业务库。
- 提交时区分普通 HTTP 写入（失败不能静默重放）与持久幂等队列（经明确去重键重试）。
- Git 源码回退使用可追溯的修复或 revert；数据库 schema、图片/对象及密钥不能仅凭回退 Git 提交恢复。
- 恢复真实业务前按 [备份与灾难恢复](runbooks/backup-restore.md) 核对迁移、SOS 请求幂等、通知 outbox 及人工接手。
- 生产联调按真实设备、通知网关和现场处置逐项验收。
