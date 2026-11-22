# 无障碍与适老化反馈 / Accessibility

岁宁面向照护机构、家属与不同技术经验的使用者。无障碍与适老化是持续改进的方向，本文件说明如何反馈实际使用障碍，不代表全部页面已通过 WCAG 2.1 AA 或第三方独立认证。

## 需要关注的问题

- 键盘无法完成操作、焦点不可见、对话框无法退出。
- 放大字号或系统缩放后，文字、按钮、表格或错误提示被遮挡。
- 屏幕阅读器无法理解表单标签、健康状态或操作反馈。
- 颜色对比不足，或动画无法按系统偏好减少。
- 地图和趋势图缺少可读的文字说明或替代操作路径。

## 反馈方式

使用 [无障碍问题表单](https://github.com/WatchTestCs/smart-elderly-watch-management/issues/new?template=accessibility_barrier.yml)。注明使用的分支或版本、页面路径、操作步骤、浏览器或 WebView2、视口和缩放比例、辅助技术（如适用），以及预期和实际结果。

请仅提交模拟或脱敏证据，不得公开真实长者的健康数据、地址、定位信息、账户信息或设备密钥。可被利用的安全问题应使用 [私密安全报告指引](.github/SECURITY.md)。

当前项目的代码开发和界面验证主要在 `preview` 进行；其 [详细无障碍设计说明](https://github.com/WatchTestCs/smart-elderly-watch-management/blob/preview/ACCESSIBILITY.md) 与本默认分支的发布功能可能不同。具体覆盖和测试结果请以对应版本的代码与验证记录为准。

---

**English:** Accessibility reports are welcome. Use the [accessibility issue form](https://github.com/WatchTestCs/smart-elderly-watch-management/issues/new?template=accessibility_barrier.yml) and include the tested version, steps, assistive technology, viewport and expected behavior. Never attach real health, location or credential data. The project does not claim completed independent WCAG certification.
