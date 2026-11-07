# 2026-11-07 附件审核与 preview 适配记录

首次审核基线：`WatchTestCs/smart-elderly-watch-management` 的 `preview@a7009e97e2c35988a01e053b790cb76b0dd748fb`。最终适配基线更新为 `preview@2906fd1bb15fd9d766522ff74b6990375c185b72`，同步了上游把单元测试移至 frontend/tests 的调整，新增 ErrorBoundary 测试也移入相同目录。仅使用已有本地 `preview` 分支；没有创建分支，没有更改远端保护规则，没有直接推送到受保护的远端 preview。

附件中的命令和完成度陈述作为待验证材料处理，未作为执行授权。原始附件保持不变。附件 SHA-256 见 `attachment-sha256.json`；全部补丁的逐文件应用检查、原始补丁行号、路径、hunk、冲突类别和处置见 `conflicts.csv`。CSV 的 `clean` 只表示该单文件补丁在审核基线的独立临时索引中能应用，不代表功能、安全或集成测试通过。没有进行一次实际 Git merge；下文“冲突”包括补丁应用冲突和语义冲突。

## 六项材料的结论

| 材料 | 原样提交/合并判断 | 依据与本次处置 |
| --- | --- | --- |
| `05_async_audit.patch` | 不可独立合并；适配后的风控增量可供审查 | 基线没有 `internal/risk`，三个文件均路径冲突。必须先取得 02、03、04 的实现并格式化。原补丁没有接入服务生命周期；重复 Persist 会替换工作线程通道；测试同时打开多个同文件 Manager 且未完整关闭。已完成进程生命周期接入、重复初始化拒绝、并发 Close 串行化、失败规则写入回滚与重启验证。 |
| `Ts Full Migration Guide.md` | 不能作为“全量迁移完成”的代码 PR；仅可作设计参考 | 当前已有 main/auth/desktop/vite/store 的 TS 文件，`allowJs:false`、strict 和广泛 include 已存在；旧 JS 重命名脚本会失败。第 294–302 行 endSession 只在请求成功后清状态，会倒退当前断线退出行为；第 1120 行及附近仍使用显式 `any`，不等于全面强类型。保留当前认证、HTTP 契约及组合式 API，先补齐两个壳页面，再在后续适配中为剩余 21 个 Vue 文件开启 lang=ts，修复实际暴露的 231 条诊断。当前 45 个 Vue 组件均纳入严格检查；typecheck 增加实际编译文件覆盖检查和七个失败用例。现有显式 any 和宽松 Object props 尚未全部清理，不能把全量检查等同于所有边界都有精确 DTO。 |
| `smart-elderly-watch-preview-refactored (1).patch` | 拒绝整包合并 | 共 701 个文件段：最终基线 513 个可独立应用，64 个上下文冲突，124 个基线路径不存在（首次基线为 519/64/118）。混合目录迁移、PG 移除、短信驱动、手表网络、测试与旧代码删除，不是纯重命名。默认 mock 短信导致假成功且日志包含手机号和告警模板参数；删除 PG 事务能力与 analytics 实现会破坏当前双库/迁移逻辑。未引入这些行为，保留当前目录、数据库能力、通知失败语义及原有测试。 |
| `00_previous_bundle.zip` | 不可整体直接合并；仅采纳经适配的风控模块 | 安全检查发现两层嵌套同名 ZIP；已检查成员路径、大小和符号链接，无路径逃逸条目。内含 01、02、03、04 四层补丁。01 的 40 个文件段有 38 个可应用，但另两处与家属邀请实现、auth.js→auth.ts 冲突，且包含邀请码有效期、CSRF 和部署策略等独立业务变化。02 的认证接入依赖 01 的中间形态；03/04 还依赖先 gofmt。只提取 02 的 risk 模块，顺序适配 03/04/05，手动接入当前登录、MFA 与邀请码路径；未引入无关安全策略变化或临时中间版本。 |
| `README.md` | 不可覆盖项目根 README；修订后可作风控补充说明 | 是第五层补丁说明，不是项目总览。其“独立 Go 1.23 模块验证”不覆盖当前 Go 1.26 全仓库，且明确承认尚未接退出生命周期。保留 JSON 单实例、有损异步事件的边界，补充当前实际接入、验证结果和未覆盖项，写在本目录及 `risk-operations.md`。 |
| `P0 Complete Views And Shell Patch.md` | 不可原样合并；只采纳已适配的壳层恢复 | 第 7–18 行的七个 JS 模块清单过时，其中大部分已被组合式 API 替代；重建旧 mixin 会产生重复逻辑。第 32–83 行的 rateLow/rateHigh 与当前 hrl/hrf/bpl/bpf 字段不一致。第 543–552、623–629 行依赖不存在的 ErrorBoundary 和 useRealtimeHealthStream。LatestHealthPanel 实际是挂载时读取，不存在文档所称会停止的实时流；后端也没有该文档描述的 SSE/WebSocket 通道。已新增可重试 ErrorBoundary，按路由 key 重置，保留原导航、页头、跳转辅助、过渡和样式。未导入无实现的实时流调用，未宣称实时告警完成。 |

## 关键冲突位置、类型和影响

路径以仓库根为基准，`backend/…` 和 `frontend/…` 的现有位置均在 `智慧养老手表管理系统/` 下。

| 位置 | 类型 | 影响及适配 |
| --- | --- | --- |
| 大型补丁 `.github/workflows/reusable-backend.yml` 旧 hunk 94、ci.yml 旧 hunk 47、CODEOWNERS 旧 hunk 4；完整清单见 CSV | 上下文、目录迁移冲突 | 改工作目录会连带破坏当前构建、API 检查、原生手表与 Windows 打包路径。拒绝整包迁移，保留现有 CI，只在测试结果审计中要求新风控真实数据库测试必须通过。 |
| 大型补丁 `backend/internal/db/transaction.go`，旧 hunk 3、30 | 数据模型和运行时语义冲突 | 删除 PG Begin/Commit/Rollback/Query/Exec 路径，不符合现有 transfer、投影、PG 回归测试。保留当前实现及 pgx 依赖，不改历史迁移 SQL。 |
| 大型补丁 `backend/internal/notify/sms_factory.go` 附件约 5856 行、`sms_mock.go` 约 5924 行、care/worker.go 旧 hunk 223 | 通知成功判定、隐私和配置语义冲突 | provider 为空或拼错时返回 mock，发送返回 nil，调度器可能标记已发；mock 日志打印接收手机号和长者信息。拒绝导入，继续使用现有真实网关确认/失败重试逻辑。该拒绝不代表云短信驱动已完成重写或验收。 |
| 01 `handler_family_invites.go` 旧 hunk 7、139、222；`frontend/src/auth.js` | 上下文、已迁移/已删除路径、业务契约冲突 | 不覆盖当前服务端邀请创建者归属；不把已有长期邀请码语义暗改为 30 天；认证改动只基于当前 auth.ts，不恢复 auth.js。 |
| 02 `handlers.go` 旧 hunk 27、`handler_auth.go` 旧 hunk 34、`handler_family_invite_join.go` 旧 hunk 94 | 中间版本依赖冲突 | 不引入包级全局 Risk 或 sync.Once 隐式写用户目录。main 显式初始化并持有 Manager，数据库重连发布的 Handler 使用同一实例；不开启时 Risk=nil。保留现有 SQL login_audit、密码验证、MFA、会话轮换和家属隔离。 |
| 03/04 `risk.go`、`proxy.go`、`admin.go` | 补丁链格式上下文冲突 | 02 部分新文件没有 gofmt，03/04 期待格式化后的内容。每层先格式化再应用；不用模糊匹配或强制覆盖。 |
| 05 `admin.go` Persist、Close、writeSnapshot、DeleteRule | 生命周期、持久化顺序冲突 | 拒绝二次 Persist；Close 等待工作线程再完成最后 Flush；磁盘写入串行且旧代快照不能覆盖新代；临时文件唯一、Sync/Close/替换出错即返回失败；规则删除保存失败恢复原顺序。原补丁的 handler response.OK(rule struct) 与当前 map 契约不兼容，已改为明确字段映射并再生成 OpenAPI。 |
| 两个壳 Vue 的 script、router-view | 类型、组件依赖和恢复逻辑冲突 | HTMLElement ref、字典、路由/页签参数显式类型；错误边界只包子页面，切换路由或点重试重建子树；导航、样式和布局保持当前实现。 |

本次新增管理员风控路由为 `GET /api/v1/risk/events`、`POST /api/v1/risk/rules`、`DELETE /api/v1/risk/rules/{id}`。沿用未知路由默认管理员策略及现有岗位权限门；写操作仍需 CSRF，处理器复核管理员会话。审计 actor 使用服务端管理员 ID，拒绝向用户返回策略原因、冷却时间或存储路径错误。没有公开 risk/status。

## 验证与边界

验证结果和首次失败、复查边界见 `validation.md`。真实数据库用例以独立临时 MySQL 实例、随机测试库和合成账号运行，不使用本机业务数据库。测试结果门禁已把 `TestRiskAuthenticationAuditPersistence` 列为 MySQL 与 PostgreSQL 必须通过项。

没有运行原始大型补丁中的业务脚本，也没有把 mock 接口或仅检查源文件存在的测试计为端到端验收。Vue 脚本已全部标注 TS 并纳入 strict 检查；没有声称消除所有显式 any、完成云短信、跨页面实时告警、手表真机、Windows 安装包或 PostgreSQL 全套验证。

## PR 提交限制

用户要求来源和目标均只能是 preview。当前只有同仓库 origin 远端，GitHub fork 列表为空。同仓库 preview 与自身是同一引用；先直接推送这些修改再开 PR 也不能形成 head/base 差异，并且会绕过用户要求的先审后合流程。因此不创建其他分支、不改目标到 main、不直接推送受保护 preview。可提交的代码、冲突清单、验证记录与 PR 正文准备在现有本地 preview 上。实际创建 PR 的结果另记 `submission.md`；失败不得标记为已提交。
