# 当前版本下载

当前公开预览版本为 **2.1.0**。安装版、免安装版和源码包均以 GitHub Release 中的附件为准。

[打开 v2.1.0 Release](https://github.com/WatchTestCs/smart-elderly-watch-management/releases/tag/v2.1.0)

| 文件 | 用途 |
| --- | --- |
| [Windows Setup](https://github.com/WatchTestCs/smart-elderly-watch-management/releases/download/v2.1.0/Suining-Setup-2.1.0-win-x64.exe) | Windows x64 安装程序 |
| [Portable ZIP](https://github.com/WatchTestCs/smart-elderly-watch-management/releases/download/v2.1.0/Suining-Portable-2.1.0-win-x64.zip) | Windows x64 免安装包，完整解压后运行 |
| [Source ZIP](https://github.com/WatchTestCs/smart-elderly-watch-management/releases/download/v2.1.0/Suining-Source-2.1.0.zip) | 与 2.1.0 标签对应的源码快照 |
| [SHA-256 清单](https://github.com/WatchTestCs/smart-elderly-watch-management/releases/download/v2.1.0/SHA256SUMS.txt) | 校验发行附件完整性 |

## 使用边界

- 当前 Release 附件基于发布清单标注的固定源码提交，不包含之后 `preview` 分支新增的提交。
- Windows 安装包不包含 MySQL 或 PostgreSQL 服务；数据库需独立部署。
- 当前 Windows 附件未做公开受信任的代码签名。
- 源码 ZIP 与 GitHub 自动生成的 Source code ZIP 可能具有不同目录结构；复现发布构建时应使用 Release 提供的命名源码附件和校验清单。
- 安装包不自动迁移旧数据库或本地图片。旧 PostgreSQL 业务库迁移须按 [双库配置指南](DUAL_DATABASES.md) 执行离线预检查与导入。

后续源码改动请从 `preview` 分支构建；不要把当前 Release 附件描述为包含分支最新提交。
