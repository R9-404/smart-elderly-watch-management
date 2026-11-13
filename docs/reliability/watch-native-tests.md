# 原生系统测试与证据

`SuiningWatch/entry/src/ohosTest` 使用官方 Hypium 1.0.25 和 TestKit；版本与 SHA-512 固定在 oh-package-lock.json5。默认产品目标不导入测试代码。测试是独立 entry_test feature 模块，没有启动器 skill 或生产调试路由。构建前运行官方 ohpm，锁文件内容漂移即拒绝，Windows 行尾差异不算依赖变化。

## 执行与核验

先按照穿戴端 README 构建 `-BuildTarget ohosTest`，再运行 `scripts/run-native-tests.ps1`。运行脚本核对构建成功、源码 dirty 状态、HAP SHA-256、bundle/module/type 和 HDC 目标后安装测试模块；不卸载、重置或清空已有应用。使用 SDK 的 `aa test` 运行框架，日志必须显示三个指定用例全部通过、0 failure/error/ignore。仅 CLI 返回 0 不够：实际失败的 Hypium 测试也可能让 aa 显示 TestFinished-ResultCode: 0。

- native_signing：生产 DeviceSigner 调用真实 CryptoArchitectureKit SHA256、变长 HMAC 密钥及安全随机源。两次签名的 nonce 不同；只输出公开固定测试输入的签名证明，由主机 .NET HMAC 独立校验原始 UTF-8 大正文，避免只检查签名长度。
- native_vault_authentication：生产 WatchVault 调用真实 HUKS，多段 UTF-8 加密解密、篡改密文拒绝，失败后原密文仍可解密。不删除、导出或替换已有 HUKS AES 密钥。
- native_store_disk_reload：生产 WatchStore 以独立名称 suining_watch_native_test_v1 保存 200 条合成采样、公开测试凭据、已验证归属、策略和 SOS，移除 Preferences 缓存后从系统持久化文件重载。移除凭据后采样与 SOS 仍保留。这是磁盘重载，不是操作系统重启。

## 存储隔离

API 24 实测 createModuleContext 不保证独立 Preferences 路径。初版测试因此触发隔离断言，在打开存储前停止；不能移除断言后直接使用生产文件。现通过显式 Preferences 名称隔离，生产默认名称 suining_watch_sim_v1 和调用方式保持原值；名称仅允许 1–64 个字母、数字、下划线或连字符。测试不得将真实数据写入测试名称，也不得清除生产文件以使测试通过。

专用模拟器验收可在本机私有证据目录读取测试前后的生产文件并仅保存摘要，确认测试没有写入它；不得将业务 XML、凭据、私钥或屏幕背后的用户内容上传至 PR。

## 证据范围

这三项测试证明系统密码服务和实际 Preferences 持久化边界，不依赖 Node SDK 替身。它们不等于完整原生网络/自动采样端到端、ArkUI 按钮生命周期或真实硬件准确度验收。真实 Go/MySQL 严格签名联调、Node 调度回归、默认 HAP 编译和当前系统测试分别记录源码 SHA 与产物摘要。Windows 原生 CI 在 SDK 配置有效时构建产品和测试包；没有模拟器镜像时只声明编译，不声明原生运行。

平台入口：[官方单元测试指南](https://developer.huawei.com/consumer/cn/doc/harmonyos-guides/unittest-guidelines/)、[Ability Assistant](https://developer.huawei.com/consumer/en/doc/harmonyos-guides/aa-tool/)。
