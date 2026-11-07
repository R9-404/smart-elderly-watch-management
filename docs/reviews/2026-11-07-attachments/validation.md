# 验证记录

最终适配基线：`preview@2906fd1bb15fd9d766522ff74b6990375c185b72`。这是工作树验证记录；未把远端旧 SHA 的 CI 当作本次新增代码的验证。最终保存的本地提交可通过该提交包含的文件树复核。原始 JSON 测试日志保存在本机审核证据目录，没有把业务密钥、会话或用户数据库复制入仓库。

| 检查 | 结果 | 覆盖与限制 |
| --- | --- | --- |
| `npm ci --ignore-scripts --no-audit --no-fund` | 通过 | 首次使用现有 lockfile；后续仅增加锁定的开发期 @types/leaflet 1.9.22 及其 @types/geojson 类型依赖，没有新增生产运行依赖。 |
| `npm run typecheck` | 通过 | 后续补齐 21 个 Vue 脚本后重新执行，覆盖 45 个组件和所有 src TS 文件；新增编译范围检查拒绝未标注 TS、恢复 JS、缩小 include、关闭严格子选项、排除源文件及未被 include 覆盖的扩展名。 |
| `npm test` | 最新 88 项通过 | 包含新增错误边界渲染失败隔离、重试恢复、导航保留、切换路由恢复；错误边界测试位于 tests/components；另七项验证严格检查覆盖门禁确实能拒绝错误配置/源文件。 |
| `npm run build` | 通过 | Vite 8.3.4。仍有 >500kB 分包体积提示，未改变告警阈值掩盖它。 |
| `go test -json -race -count=1 -timeout 180s ./...` | 1021 项通过、215 项跳过、无失败 | 本轮未提供数据库/S3环境；跳过主要为真实数据库、S3、专用压测等，不能计作通过的集成测试。 |
| 独立 MySQL 8.0.43 风控复查 | 四个顶层测试通过，无跳过/失败 | TestRiskGatesBeforeDatabaseAccess、TestRiskManagementRequiresAdminAndCSRF、TestRiskAuthenticationAuditPersistence、TestServerRiskOptInAndInitializationFailure。新增会话主体校验后独立复查仍通过，新认证/持久化用例约 75 秒，含完整新库迁移；真实 SQL 成功登录审计、冷却、例外规则、会话和补充 JSON 重启加载都经断言。 |
| 独立 PostgreSQL 16 风控及登录兼容复查 | 五个顶层测试及两个登录角色子用例通过，无跳过/失败 | 包含风控认证、JSON 重启、原有登录响应契约与管理员/家属会话。初次失败发现测试误查 MySQL 专用 login_audit 表；按现有 loginAuditTx 仅在 MySQL 记权威 SQL 审计的边界修正测试，保留 PG 行为，未增加假表或跳过风控断言。 |
| 首次 MySQL 测试 | 失败，原始记录保留 | 总 120 秒到期时第二个用例仍在迁移阶段，并非断言通过。旧登录响应契约测试及其管理员/家属子用例已通过。缩小到所需测试范围并将总限时设为 240 秒后复查通过，没有通过重试业务断言或关闭持久化来变绿。 |
| Go build/vet、Go 格式检查 | 通过 | 359 个 Go 源文件按 LF 标准检查；不批量改写历史 SQL 或无关源文件。 |
| OpenAPI 生成与 `--check` | 通过 | 112 个路径；新增 DELETE 合约明确路径 id、CSRF 且无需 JSON body，有物化 schema 回归断言。 |
| 原生 Go API 边界检查、`git diff --check` | 通过 | 仅检查当前代码/补丁卫生，不能代替运行时验收。 |
| Python reliability 单元测试 | 16 项通过 | 保留原有发布/监测测试。新数据库用例已经加入 MySQL 和 PG 结果审计的必过名单。 |

风控额外断言包括：慢磁盘写锁不阻塞事件入队、并发 Close、重复 Persist 拒绝、旧快照不能回盖新规则、规则保存失败恢复原顺序、失败修改不会在后续 Flush 或重启中复活、临时文件清理、管理员身份和 CSRF 校验、风控拒绝不泄露策略/原因、损坏的策略文件不静默加载空规则。

本次只补跑了 PostgreSQL 风控与登录兼容子集，没有实跑 PostgreSQL/S3全套、Go 1.27.2、远端完整 CI、Windows 打包/安装、鸿蒙真机或云短信。代码仍应经过这些适用门禁及独立审查后才可判定可合并；不得把本地通过称作全部验收完成。

HTTP 正常 Shutdown 完成后再关闭后台和风控。若既有 10 秒排空期限到期，会记录错误并关闭连接；这种异常退出不保证所有在途补充事件完整。JSON 本身也不提供崩溃零丢失保证，详见 risk-operations.md。

后续验证仍为本地工作树验证，不能把日志里的基础 HEAD 当作包含未提交变更的远端 CI SHA。后续结果、原始日志摘要和被验证文件摘要见 followup-validation.json。PostgreSQL 启动包装器曾因 Windows 子进程继承输出管道而等待超时；已确认服务实际存活，在同一实例完成测试后按所属数据目录关闭，没有因观察超时重启实例。

## 用户授权 fork 后的本轮复查

上游 preview 仍为 2906fd1，基于该版本完成小补丁适配及事件缓冲优化；大型重构补丁暂缓。此次 Go 全包竞态检查为 1023 项通过、206 项跳过、无失败，未配置数据库/S3环境；前端仍为 88 项通过，45 个 Vue 组件严格检查和生产构建通过。Go build/vet、OpenAPI 112 路径、API 边界 864 个跟踪路径及 Python 16 项均通过。

另外启动独立的临时 MySQL 与 PostgreSQL 实例，以竞态模式执行风控集成子集：MySQL 四个顶层用例通过；PostgreSQL 五个顶层用例及两个登录角色子用例通过；均无跳过或失败。正常完成后两个测试实例均已停止，不使用业务数据库。完整记录及日志摘要见 fork-validation.json。

新增测试覆盖满容量保留顺序、导出历史结果与在途快照独立性，以及过期事件在正常刷盘重启后不再出现。纯内存基准三轮分配量从 548880 B/op 降至 1053–1054 B/op；不以此推算磁盘、SQL或生产性能。仍有大于 500kB 的前端分包告警；尚未实跑远端全套 CI、PG/S3 全套、真机、云短信或 Windows 安装验收。

## TS 6 / Vue 升级追加验证

用户授权把新 TS 6 指南适配到 PR #106 后，前端升级至 TypeScript 6.0.3、Vue 3.5.45；97 项测试、三个项目的普通及 --stableTypeOrdering 检查、Vite 构建均通过。CI 已加入稳定排序命令。锁定安装 npm ci 通过，依赖完整性检查解析到指定版本；审计全部依赖为 high=0、critical=0、moderate=5，五项属于未改变版本的测试工具链；生产依赖审计为零。没有执行 npm audit fix --force。

新比较覆盖 45 个组件的模板、样式及编译后的 Props/emits/expose AST，忽略换行、引号及位置后均与升级前 5b004aa 一致。检查门禁共 14 项回归，另有 RecordCards 的 props 替换与 actions 插槽行为测试。现有隐式/显式边界并未因此等同于所有运行时路径校验。

详细新阶段记录为 ts6-validation.json。后端工作树相对 5b004aa 无改动，不重复以同一源码进行数据库验收；沿用此前实跑结果并等待当前 PR head 的 CI。原 fork-validation.json 是小补丁阶段记录，其前端源码指纹不能用来证明后续 TS 6 新提交。
