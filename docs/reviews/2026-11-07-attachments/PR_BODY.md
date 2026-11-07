本 PR 基于上游 `preview@2906fd1bb15fd9d766522ff74b6990375c185b72`，适配小补丁中的异步风控审计、页面错误恢复及 TypeScript 6.0.3 / Vue 3.5.45 升级。来源仅为 `R9-404:preview`，目标仅为 `WatchTestCs:preview`。

## 变更

- 从 `00_previous_bundle.zip` 的风控增量顺序适配至 `05_async_audit.patch`，接入显式 opt-in 的进程生命周期；修复重复初始化、并发关闭、规则失败回滚及旧快照覆盖，沿用已有登录、MFA、CSRF、岗位与家属隔离。SQL 登录审计和 PG 兼容路径保持原有边界。
- 优化满容量事件缓冲区，避免每次追加都复制 2000 条记录；补充保留期限、顺序和快照隔离回归。在本机纯内存基准中分配量从 548880 B/op 降至约 1054 B/op，不作为生产性能结论。
- 按 P0 文档适配可重试 ErrorBoundary，保留两个壳的导航、页头、辅助跳转和样式；切换路由可恢复。未调用文档中缺失的实时流接口。
- 按 TS 指南为剩余 21 个 Vue 脚本补齐类型标注及实际诊断修复，45 个组件均纳入严格检查；增加 DTO、表单、地图类型和编译覆盖门禁测试。已有显式 any 尚未全部清理。
- 审核新 TS 6 指南后锁定 TS 6.0.3、Vue 3.5.45 和官方 Vue preset；移除 baseUrl，分离应用/Node/测试配置，普通及稳定排序检查均覆盖三个项目并接入 CI。响应式 Props、useTemplateRef 和元组 emits 保留原契约；45 个组件模板、样式及编译后的公开结构比较一致。纠正文档混淆 asserts 和导入断言等建议，未引入新的原生 API 运行要求。
- 新增管理员风控 API、OpenAPI 合约及逐项附件审核、冲突位置与适配依据；README 仅作为风控补充说明。

## 明确排除

按用户本轮要求，约 13 MB 的 `smart-elderly-watch-preview-refactored (1).patch` 暂缓，不包含其目录迁移、PG 删除、默认 mock 短信或手表重写。ZIP 中无关的邀请有效期等策略变更也未引入。JSON 风控仍是单实例、有损的补充信号；未宣称实时告警、精确 DTO 全覆盖、真机或云短信验收完成。

## 验证

结果见 `docs/reviews/2026-11-07-attachments/validation.md` 、小补丁 `fork-validation.json` 及升级 `ts6-validation.json`（两者验证阶段不同）。TS 6 升级后本地前端 97 项测试、45 组件普通/稳定排序严格检查及构建通过；Go 全包竞态测试 1023 项通过、206 项因缺少数据库/S3/专用场景配置而跳过。Go build/vet、112 路径 OpenAPI、API 边界及 16 项 Python 测试通过。独立 MySQL 风控四个顶层用例、PostgreSQL 风控及登录兼容五个顶层用例和两个角色子用例均在竞态模式下通过，无跳过；测试实例已停止。远端 CI 以 PR 当前提交的 checks 为准，本地未执行 PG/S3 全套、Go 1.27.2、真机、云短信或 Windows 安装验收。

详细审核与逐文件冲突清单：`docs/reviews/2026-11-07-attachments/README.md`、`conflicts.csv`。

TS 6 适配细节和逐项建议处置见 `docs/reviews/2026-11-07-attachments/ts6-upgrade.md`。依赖审计 high/critical 为零，原有测试工具依赖链保留五项 moderate；未宣称依赖完全无告警。
