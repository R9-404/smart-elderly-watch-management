# PR 提交记录

## 当前授权及来源

2026-11-07 用户明确授权在个人账号创建 fork，接受 GitHub 自动复制上游已有分支，并要求只适配其余小补丁，暂缓约 13 MB 的大型补丁。已通过 GitHub API 创建并核验：

- fork：`https://github.com/R9-404/smart-elderly-watch-management`
- parent：`WatchTestCs/smart-elderly-watch-management`
- head：`R9-404:preview`
- base：`WatchTestCs/smart-elderly-watch-management:preview`
- 适配基线：`2906fd1bb15fd9d766522ff74b6990375c185b72`

仅在本地已有 preview 工作；fork/preview 为 GitHub 自动复制的已有分支，没有手动新建功能分支，没有直接推送上游或修改保护规则。原始附件、用户便携版和本机运行脚本均未纳入提交。大型补丁的历史审核资料保留，但其重构代码未导入。

本文件先记录创建 fork 与待提交方向；PR 创建后的编号、URL、实际 head/base 复核结果将在成功后追加。PR 正文为 PR_BODY.md；本地验证见 validation.md 和 fork-validation.json。不得将提交 PR 等同于完成远端 CI 或全部产品验收。

## 历史：同仓库尝试

此前按用户原先禁止新分支及 fork 的约束，向同仓库发起 head=preview、base=preview 的草稿 PR 请求。GitHub 返回 HTTP 422：`No commits between preview and preview`，当时 fork 列表为空，没有创建 PR，也未直接推送上游。后续用户明确授权 fork 后，该阻碍已通过上述跨仓库来源解除。
