# 文档索引与维护边界

此处集中列出 **`preview` 当前代码的操作文档**。发布版本的实际二进制、提交、发布清单和 SHA-256 以其 Release 附件为准；`preview` 文档更新不会改变旧的发行资产。

## 从哪里开始

| 需求 | 唯一权威入口 |
| --- | --- |
| 安装与快速运行 | [项目 README](../README.md) / [English](../README.en.md) |
| 模块边界、部署关系、会话和故障模型 | [实现架构](ARCHITECTURE.md) |
| 前后端 API、路径和字段真相源 | [接口契约](API_CONTRACTS.md)（Go 路由、`contracts.json`、OpenAPI） |
| Windows 构建与签名 | [构建指南](BUILD.md) |
| 长期独立照护服务及通知队列 | [照护服务部署](CARE_SERVICE_DEPLOYMENT.md) |
| MySQL 主库、可选 PG 分析和旧库迁移 | [双库操作](DUAL_DATABASES.md) |
| 真实数据库、前端、手表和桌面测试范围 | [测试分层](TESTING.md) |
| HarmonyOS 手表开发/协议/模拟器 | [穿戴端 README](../SuiningWatch/README.md) |
| 界面可用性与辅助访问 | [UI / UX 规范](UI_UX.md) |
| 社区问题反馈及脱敏 | [支持与反馈](SUPPORT.md) |

## 运行、安全和恢复

- [MySQL 和文件备份/隔离恢复](runbooks/backup-restore.md)
- [PG 可选分析、故障注入和独立告警](runbooks/optional-pg-monitoring.md)
- [图片对象、多实例、GC 及协调快照](reliability/storage-operations.md)
- [滥用控制、限流、Alertmanager 和独立值守](reliability/abuse-and-monitoring.md)
- [数据分类、密钥轮换和历史明文风险](reliability/data-protection.md)
- [可追溯发布、分支保护及回滚](reliability/release-governance.md)
- [性能测量和原生优化的复现](reliability/performance.md)

## 不要混用历史记录与当前合同

- `docs/releases/` 是**已发布版本的历史记录**，不能随现行功能改动而覆盖历史结论。
- `evidence/` 保存具备来源 SHA 的实验日志或原始证据；不直接删除/重写原始字节或哈希清单。只清理重复的人工总结，保留原始材料与复现方法。
- `docs/licenses/`、字体及图标许可证、第三方版权说明为必要法律材料，不属于冗余说明。
- 阶段性“任务完成”“准备合并”“本轮 CI 全绿”不得放进无版本的长期手册；请在对应 PR、发行清单或 SHA 绑定的证据中记录。
- 不复制所有路由、数据字段和配置默认值到多份 Markdown。路由及 schema 从 Go/生成器读；版本从 `desktop/version.json` 读；测试命令及版本从 `.github/workflows/` 和对应工具配置读。
- 修改或删除文档必须同时修正入站链接，并区分：保留运行手册、归档历史验收、移除仅对单次任务有效的流水账。不能为了减小文件数而删掉恢复演练、紧急照护和隐私要求。
