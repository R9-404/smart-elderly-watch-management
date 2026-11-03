# 测试、验证与证据分层

本文说明如何复现当前 `preview` 的验证范围。**一次测试结论只适用于实际执行的完整 Git SHA、数据库类型、操作系统和工具链**；旧 Release 或 PR 的绿色记录不能直接转用于新提交。

## 变更与最小验证范围

| 变更类型 | 最少验证 | 需要增加的验证 |
| --- | --- | --- |
| 仅文档或链接 | 文件路径、内部链接、代码示例与当前源文件一致 | 若调整了 OpenAPI 示例，运行 `python3 scripts/generate-openapi.py --check` |
| Vue 页面、API 类型和状态 | `npm test`、`npm run typecheck`、`npm run build` | 页面交互变化运行对应 Playwright 夹具，验证取消、错误提示和空状态 |
| Go 业务逻辑 | 相关包测试、`go test ./...`、`go vet ./...` | 事务、设备权限、队列和竞态调整需真实 MySQL / PG 与 `-race` |
| 设备/手表协议 | 模拟器与 ArkTS 转译协议测试 | HAP 原生编译、模拟器运行、真机传感器、绑定代次、断网重试分别验证 |
| 数据库迁移或备份恢复 | 独立测试库升级和回滚/恢复路径 | 正式操作前按数据体量做隔离恢复、校验和、写入暂停和签名密钥核对 |
| Windows 安装或签名 | 原生打包、sidecar 哈希、启动及卸载检查 | WebView2 缺失补装、真实桌面交互、目标系统与证书链 |

不要为了文档变更额外实施数据库破坏性演练。当前必需 CI 检查和命名以 [工作流](../.github/workflows/ci.yml) 及该提交的实际 check-runs 为准。

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

上面的 900 秒仅为本地回归建议；**CI 的实际命令和限制以 [reusable-backend.yml](../.github/workflows/reusable-backend.yml) 为准**。CI 还对 PG 兼容包单独执行一次配对验证。跳过的数据库测试不能记为通过，所有输入必须指向隔离环境，禁止运行在生产库。

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

Windows 桌面构建、WebView2 和签名边界见 [BUILD](BUILD.md)。代码/安装包的来源必须绑定完整 SHA、真实构建记录、版本及哈希。安全、质量和原生构建结果应分别报告通过、失败、跳过、尚未运行；CI 绿灯不代表外部通知已送达或值班人员已经接单。

## 数据、安全与恢复

- 自动化测试和故障注入只在独立克隆、临时库、测试对象存储及假网关运行，不得覆盖运行中的业务库。
- 提交时区分普通 HTTP 写入（失败不能静默重放）与持久幂等队列（经明确去重键重试）。
- Git 源码回退使用可追溯的修复或 revert；数据库 schema、图片/对象及密钥不能仅凭回退 Git 提交恢复。
- 恢复真实业务前按 [备份与灾难恢复](runbooks/backup-restore.md) 核对迁移、SOS 请求幂等、通知 outbox 及人工接手。
- 未部署的外部短信/语音网关、真实穿戴设备和生产现场处置流程，不得写成已通过。
