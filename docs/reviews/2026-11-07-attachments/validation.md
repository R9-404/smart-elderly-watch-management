# 验证记录

最终适配基线：`preview@2906fd1bb15fd9d766522ff74b6990375c185b72`。这是工作树验证记录；未把远端旧 SHA 的 CI 当作本次新增代码的验证。最终保存的本地提交可通过该提交包含的文件树复核。原始 JSON 测试日志保存在本机审核证据目录，没有把业务密钥、会话或用户数据库复制入仓库。

| 检查 | 结果 | 覆盖与限制 |
| --- | --- | --- |
| `npm ci --ignore-scripts --no-audit --no-fund` | 通过 | 使用现有 lockfile，没有增加前端依赖。 |
| `npm run typecheck` | 通过 | 最终上游测试目录调整后重新执行；并不意味着未写 lang=ts 的 21 个 Vue 文件已完成严格检查。 |
| `npm test` | 81 项通过 | 包含新增错误边界渲染失败隔离、重试恢复、导航保留、切换路由恢复；最终测试位于 tests/components。 |
| `npm run build` | 通过 | Vite 8.3.4。仍有 >500kB 分包体积提示，未改变告警阈值掩盖它。 |
| `go test -json -race -count=1 -timeout 180s ./...` | 1021 项通过、215 项跳过、无失败 | 本轮未提供数据库/S3环境；跳过主要为真实数据库、S3、专用压测等，不能计作通过的集成测试。 |
| 独立 MySQL 8.0.43 风控复查 | 四个顶层测试通过，无跳过/失败 | TestRiskGatesBeforeDatabaseAccess、TestRiskManagementRequiresAdminAndCSRF、TestRiskAuthenticationAuditPersistence、TestServerRiskOptInAndInitializationFailure。新认证/持久化用例约 74 秒，含完整新库迁移；真实 SQL 成功登录审计、冷却、例外规则、会话和补充 JSON 重启加载都经断言。 |
| 首次 MySQL 测试 | 失败，原始记录保留 | 总 120 秒到期时第二个用例仍在迁移阶段，并非断言通过。旧登录响应契约测试及其管理员/家属子用例已通过。缩小到所需测试范围并将总限时设为 240 秒后复查通过，没有通过重试业务断言或关闭持久化来变绿。 |
| Go build/vet、Go 格式检查 | 通过 | 359 个 Go 源文件按 LF 标准检查；不批量改写历史 SQL 或无关源文件。 |
| OpenAPI 生成与 `--check` | 通过 | 112 个路径；新增 DELETE 合约明确路径 id、CSRF 且无需 JSON body，有物化 schema 回归断言。 |
| 原生 Go API 边界检查、`git diff --check` | 通过 | 仅检查当前代码/补丁卫生，不能代替运行时验收。 |
| Python reliability 单元测试 | 16 项通过 | 保留原有发布/监测测试。新数据库用例已经加入 MySQL 和 PG 结果审计的必过名单。 |

风控额外断言包括：慢磁盘写锁不阻塞事件入队、并发 Close、重复 Persist 拒绝、旧快照不能回盖新规则、规则保存失败恢复原顺序、失败修改不会在后续 Flush 或重启中复活、临时文件清理、管理员身份和 CSRF 校验、风控拒绝不泄露策略/原因、损坏的策略文件不静默加载空规则。

本次没有实跑新增代码的 PostgreSQL/S3全套、Go 1.27.2、远端完整 CI、Windows 打包/安装、鸿蒙真机或云短信。代码仍应经过这些适用门禁及独立审查后才可判定可合并；不得把本地通过称作全部验收完成。

HTTP 正常 Shutdown 完成后再关闭后台和风控。若既有 10 秒排空期限到期，会记录错误并关闭连接；这种异常退出不保证所有在途补充事件完整。JSON 本身也不提供崩溃零丢失保证，详见 risk-operations.md。
