# 文件职责导航

当前 preview 源码树包含 **994 个文件**，其中 `docs/file-descriptions.json` 保留 **386 条有效路径说明**。此计数按当前 Git 树重新计算。

| 目录或入口 | 文件数 |
| --- | ---: |
| `.github` | 27 |
| `根目录` | 8 |
| `智慧养老手表管理系统` | 639 |
| `deploy` | 13 |
| `docs` | 40 |
| `evidence` | 211 |
| `scripts` | 23 |
| `SuiningWatch` | 32 |
| `tests` | 1 |

## 当前文档入口

| 主题 | 说明 |
| --- | --- |
| [项目说明](../README.md) | 当前版本、安装、部署与构建 |
| [接口合同](API_CONTRACTS.md) | 当前请求、响应、状态码和差异处理 |
| [手表接入](WATCH_INTEGRATION.md) | HarmonyOS 通信、授权和联调 |
| [Go 后端](../智慧养老手表管理系统/backend/README.md) | 业务接口、数据库迁移和后台任务 |
| [Vue 前端](../智慧养老手表管理系统/frontend/README.md) | 管理端与家属端网页 |
| [手表工程](../SuiningWatch/README.md) | ArkTS 应用、传感器和本地采样队列 |
| [职责数据](file-descriptions.json) | 按完整仓库路径查找当前模块说明 |
| [当前发布说明](releases/2.1.0.md) | 当前预览版本的范围与限制 |
| [测试指南](TESTING.md) | 当前前后端、手表和桌面测试入口 |
| [运维手册](runbooks/backup-restore.md) | 备份与恢复操作 |

`evidence/` 中的历史结果按其记录的源码 SHA 追溯；它们不是当前提交的 CI 通过证明。当前提交验证以 GitHub Actions 对应提交的运行结果为准。
