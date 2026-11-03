# 仓库和发布保护

GitHub 分支保护由仓库管理员配置和核验。
正式启用前由仓库管理员核对目标分支、检查名称和来源，并确认规则实际生效。

建议保护规则至少包括：禁止未经审查的直接推送、force push 和分支删除；要求独立 reviewer；
最后推送者不能自行批准；新提交使旧批准失效；合并前必须通过当前 SHA 对应的必需检查。
字段说明见 [GitHub rules REST 文档](https://docs.github.com/en/rest/repos/rules#create-a-repository-ruleset)。
GitHub Actions 检查来源与名称应以仓库实际 check-runs 为准，不应硬编码未经复核的 integration ID。
CodeQL 或依赖审查因 fork 权限、配置缺失而未运行时，应由维护者在可信分支重新执行，不能把跳过当成通过。

## 预览交付

预发布只面向 `preview`，不自动更新 `main`。附件、源码快照和校验清单必须对应同一版本与提交；旧 Release 不包含后续分支改动。普通 PR/CI 不读取发布私钥，签名与发布只在明确的 release 流程执行。

## 当前发布门槛

Preview CI 与 Security 工作流分别验证当前提交的后端、数据库、race、迁移与恢复、S3、
OpenAPI、前端 test/typecheck/build/audit、Windows sidecar 及漏洞扫描。具体执行范围以当前
workflow 文件和对应 commit 的 job 结果为准；未配置的数据库测试、跳过的原生 SDK 编译、
未进行的真实设备或生产环境演练不得报告为通过。

发布门禁应校验明确的 40 位 commit SHA 与目标分支 head 一致，防止验证期间分支移动；选择
该 SHA 对应的最新工作流运行，拒绝用旧的成功运行覆盖新提交的失败/取消结果；同时核对聚合
job、SHA 命名 artifact、有效期和 SHA-256 digest。Release preparation 与 Windows installer
入口必须固定 checkout 到已验证 SHA，产物上传前再次核对来源。没有完整证据不得进入打包。

分支保护与发布权限须在 GitHub Rulesets 中单独启用。
仓库管理员必须独立启用分支规则、限制发布权限，并定期审查工作流来源和配置。

## 来源和依赖

CI artifact 应保存 source SHA、Go 版本与平台、原始测试日志、逐项 skip、构建/vet、前端 audit
和原生 sidecar SHA-256。GitHub artifact digest 是归档完整性证据，sidecar 文件 hash 是二进制
完整性证据，两者用途不同。正式安装包还需发布清单、发布 SHA 和构建 run 链接。

签名密钥不进入仓库；未签名提交或附件不得描述为已签名。生产发布还需维护者配置受托签名
tag、代码签名证书、时间戳服务、证书轮换、构建证明和留存策略。CI action、容器基础镜像、
SDK 与依赖来源应固定版本或摘要，并按补丁策略持续审核。


## 源码回退与数据回滚

- 未合并的修改通过新修复提交或撤回 PR 处理；已合并的共享分支使用可审计的 `git revert`，不要对协作分支强制推送、重写历史或删除其他人的提交。
- 回退源码前先核对当前分支与最新写入的业务 schema，**Git 回退不自动撤销数据库迁移、S3 图片格式或加密密钥**。必须确认旧二进制能读取当前 schema/密文，再在隔离副本演练回滚脚本。
- 生产恢复需要停止旧写入者和通知发送者，保护 MySQL、对象、密钥和 outbox；启用恢复实例前核对 SOS `requestId` 与网关幂等状态，避免重复呼叫真实人员。具体步骤见[备份与恢复](../runbooks/backup-restore.md)和[图片/多实例操作](storage-operations.md)。
- 性能、通知网关、真机及受保护分支管理员规则的外部验收必须单独记录，不因 CI 通过而自动视为生产通过。
