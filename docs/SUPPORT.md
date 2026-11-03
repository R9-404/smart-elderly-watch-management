# Support and feedback / 支持与反馈

[简体中文 README](../README.md) · [English README](../README.en.md) · [Contributing](../CONTRIBUTING.md)

Use [GitHub Issues](https://github.com/WatchTestCs/smart-elderly-watch-management/issues) for reproducible non-sensitive bugs and feature requests. Chinese and English are welcome. 

普通缺陷和功能建议可提交到 Issues。安全漏洞、真实人员资料、数据库文件、连接密码、设备密钥及私钥不要公开；请先阅读 [Security Policy](../.github/SECURITY.md)。

## Quick checks / 提交前排查

| Symptom | Check first |
| --- | --- |
| Database warning / 数据库未连接 | Confirm MySQL readiness, the configured DSN and completed migrations. PostgreSQL failures delay analytics but should not block core business. |
| Cannot create or use an account / 无法初始化或登录 | There is no default password. First-admin setup requires the native desktop flow and a ready database. |
| Portable app will not start / 免安装无法启动 | Extract the whole ZIP, keep both EXEs together, and confirm WebView2 is available; Setup can install the missing shared runtime. |
| Old records are missing / 升级后没有旧记录 | Back up first. Legacy PostgreSQL business data requires explicit offline import into an empty MySQL target; uploaded images are copied separately. No automatic legacy database conversion is performed. |
| Linux differs from Windows / Linux 行为不一致 | Windows x64 is the documented desktop build target. |
| Unknown publisher / 未知发布者 | The current Windows preview artifacts are not signed by a publicly trusted certificate. Treat publisher warnings as expected and verify the downloaded SHA-256. |

Logs and screenshots must be redacted: remove names, phone numbers, locations, health records and credentials.

Current Windows and source artifacts are listed in the [project downloads](../README.md#下载). New health uploads returning a binding conflict must retain the binding version from acquisition time; do not relabel old records. Notification delivery is separate from care verification.

## Bug report template / 缺陷报告模板

Copy this into a [new issue](https://github.com/WatchTestCs/smart-elderly-watch-management/issues/new):

```text
Summary / 问题概述:

Version or commit / 版本或提交:
Exact package filename / 安装包名称:
OS and architecture / 系统与架构:
Distribution / 交付方式: Windows Setup / portable ZIP / source build / other (state precisely)

Steps to reproduce / 最小复现步骤:
1.
2.

Expected / 预期行为:
Actual / 实际行为:

Database state / 数据库是否已连接:
Frequency / 是否稳定复现:
Redacted logs or screenshots / 脱敏证据（可选）:
```



## Feature request template / 功能建议模板

```text
Use case / 使用场景:
Who is affected / 使用者:
Current limitation / 现有限制:
Desired outcome / 期望结果:
Alternatives considered / 已考虑的替代方法:
How to verify success / 如何验收:
```

Describe the user need before prescribing a framework or dependency. For database or authentication changes, explain compatibility and whether existing data could be affected.


