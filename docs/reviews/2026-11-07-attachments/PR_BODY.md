本次基于 preview 审核六项附件，移植经过适配的风控链及子视图错误恢复，并补齐 Vue 脚本严格检查覆盖。风控为显式 opt-in，保留 SQL 登录审计、MFA、会话轮换、家属权限及当前目录，修复异步写盘退出接入、重复初始化、规则失败回滚、旧快照覆盖、管理员身份与 CSRF，并补齐 OpenAPI。两个壳页面增加 TS 类型与可重试错误边界，其余 21 个 Vue 脚本也纳入严格检查，增加准确的 DTO、表单和地图类型与编译覆盖门禁，保留现有导航、页头、辅助跳转与样式。

大型 701 文件补丁未整体导入：其目录迁移与当前实现冲突，默认 mock 短信可能假成功，并删除现有 PG 能力。完整 TS 与实时流文档也不满足原样合并条件。详见 docs/reviews/2026-11-07-attachments/README.md 与 conflicts.csv。

验证：88 项前端测试、vue-tsc、Vite build；Go 全包竞态测试 1021 项通过，215 项因数据库/S3/专用场景未配置而跳过；独立 MySQL 风控四个顶层测试通过、无跳过。PostgreSQL 风控/登录兼容子集通过且无跳过；保留旧 PG 兼容路径无 login_audit 表的边界。Go build/vet、OpenAPI、API 边界及 16 项 Python 测试通过。首次 MySQL 总限时失败已保留，缩小所需范围并调整总限时后复查通过。

全量 Vue 脚本检查已覆盖，但尚未消除已有显式 any；未据此宣称精确 DTO 全覆盖、实时告警、PG/S3全套、真机、云短信或远端全套 CI 完成；JSON 风控是单实例有损补充信号，不能替代 SQL 权威审计。仅允许 head=preview、base=preview，实际创建结果见 submission.md。
