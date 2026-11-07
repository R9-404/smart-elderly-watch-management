# PR 提交记录

## 当前授权及来源

2026-11-07 用户明确授权在个人账号创建 fork，接受 GitHub 自动复制上游已有分支，并要求只适配其余小补丁，暂缓约 13 MB 的大型补丁。已通过 GitHub API 创建并核验：

- fork：`https://github.com/R9-404/smart-elderly-watch-management`
- parent：`WatchTestCs/smart-elderly-watch-management`
- head：`R9-404:preview`
- base：`WatchTestCs/smart-elderly-watch-management:preview`
- 适配基线：`2906fd1bb15fd9d766522ff74b6990375c185b72`

仅在本地已有 preview 工作；fork/preview 为 GitHub 自动复制的已有分支，没有手动新建功能分支，没有直接推送上游或修改保护规则。原始附件、用户便携版和本机运行脚本均未纳入提交。大型补丁的历史审核资料保留，但其重构代码未导入。

## 创建成功与远端复核

已创建草稿 PR [#106](https://github.com/WatchTestCs/smart-elderly-watch-management/pull/106)：`fix(preview): adapt small patches and optimize audit buffering`。

- GitHub 返回 state=open、draft=true；没有执行合并。
- head 为 `R9-404:preview`，base 为 `WatchTestCs:preview`。
- 创建时 head 为 `d25a1797661b900379aeb7128727e49f18274288`，base 为 `2906fd1bb15fd9d766522ff74b6990375c185b72`，5 个本地提交、68 个变更文件。
- GitHub 随后的实际合并检测为 mergeable=true，无文本合并冲突；mergeable_state=blocked，草稿、检查或审查门禁仍适用，不能据此宣称可立即合并。
- Preview CI、Desktop Contract 已实际启动，Security 首次检查成功；最终状态以当前 PR head 的 checks 为准，旧 head 的结果不能替代新提交验证。

5b004aa 当时仅更新提交记录，前后端 648 个已跟踪文件的 index 指纹与该阶段 fork-validation.json 中的本地测试对象一致。后续用户授权的 TS 6 / Vue 升级将继续推送到同一 PR，新增验证另见 ts6-validation.json 与 ts6-upgrade.md。PR 正文为 PR_BODY.md；本地验证见 validation.md 和 fork-validation.json。提交 PR 不等于远端 CI 或全部产品验收通过。


## 历史：同仓库尝试

此前按用户原先禁止新分支及 fork 的约束，向同仓库发起 head=preview、base=preview 的草稿 PR 请求。GitHub 返回 HTTP 422：`No commits between preview and preview`，当时 fork 列表为空，没有创建 PR，也未直接推送上游。后续用户明确授权 fork 后，该阻碍已通过上述跨仓库来源解除。
