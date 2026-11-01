# HarmonyOS 无真机验证

基线 `cfd7b418e195c860dcf490bd087a7b1f850b276d`。用户暂时没有手表，要求完成当前可执行的最大范围；私有 SDK/实现仍在允许范围内。本文不把模拟数据或 Node 转译当作原生传感器证明。

## 本轮实现

- `HealthSensors.ets` 增加前台会话与暂停/恢复。`Index.ets` 在设备认证前取得会话票据，页面隐藏、应用进入后台和组件销毁时取消读取；即使认证很慢，旧请求也不能在重新打开后启动硬件监听。只在页面销毁时取消，无法覆盖应用退到后台的情况。[ArkUI 生命周期合同](https://github.com/openharmony/docs/blob/master/zh-cn/application-dev/reference/apis-arkui/arkui-ts/ts-custom-component-lifecycle.md)。
- `tests/harmony.test.cjs` 增加延迟认证、授权期间离开回归。当前 18 项便携检查通过；使用 SDK 边界替身，未启动 ArkUI 运行时。
- `tests/backend.e2e.cjs` 增加真实回环 HTTP 故障代理：先把生产 WatchApi 的请求交给真实 Go 后端并等待数据库提交，再断开连接丢失确认。随后重新创建生产 WatchStore、从磁盘支持的 Preferences 边界适配器恢复原记录，重试须得到两条 duplicates，收到完整确认后才清空。不是用预设成功响应代替真实数据库，也不是实际手表闪存/掉电验证。本机独立 MySQL 空库完整联调 9 项通过。
- `.github/workflows/reusable-watch.yml` 增加独立 Go/MySQL 联调任务，重复执行上述故障恢复。仅使用每个 CI 容器的空测试库，没有真实通知收件人。
- `scripts/build.ps1` 支持 DevEco Studio 或独立 Command Line Tools，检查依赖及 SDK，保留环境变量、拒绝覆盖不同来源插件，先 clean 再编译，输出源码 SHA、dirty 状态、构建状态、HAP 大小及 SHA-256。缺工具明确记为 blocked，编译失败记为 failed，不用旧 HAP 充当本轮成功证据。
- Windows CI 强制执行无 SDK 失败预检，核对 blocked/null artifact 和调用方环境恢复。本轮预检真实发现“未设置变量被恢复成空字符串”的差异，已改为对原本不存在的变量删除对应进程环境项；它不代表 SDK 编译成功。
- 在现有 `Preview CI` 手动入口增加 sdk_url、sdk_sha256 参数。提供 SDK HTTPS 地址及预期 SHA-256 后执行 Windows 原生构建；校验通过才解压运行，成功时保留未签名 HAP 和构建报告。普通 PR CI 强制执行便携检查及真实 Go/MySQL 联调，原生 SDK 任务明确跳过，不将“没有 SDK”包装成原生编译通过。
- `SuiningWatch/oh-package.json5` 对齐原生工程 0.2.0 元数据。

## 私有实现复核

用户提供的原 APK SHA-256 为 `6e693fe6051b69fefa68215ce52d1f3ae4a2a7c8e25d09b034c73c20bf6f42fd`，已在本机重新核对。此前九份类签名之外，使用本机 JADX CLI 1.5.6 静态读取了 WearEngineService、WearEngineExtendService、P2pManagerImpl、MonitorManagerImpl、SensorManagerImpl；没有执行 APK，没有把反编译实现复制进产品。

P2P 实现涉及源包名/UID、对端指纹、DEVICE_MANAGER 权限、回调与 Binder 死亡；Monitor 按监控项区分权限；SensorManager 具有实际传感器枚举、异步读取与停止、SENSOR/MOTION_SENSOR 权限及设备有效性处理。这些比接口名称更接近可用接入合同，但仍属于 Android 手机侧。当前没有手机伴随端与手表侧私有服务合同，不能把这些 Binder 对象直接导入 ArkTS。

原报告“存在 ClientToken 就证明第三方必定因华为签名被拒”的推断不足：本次读取的代码能证明令牌、应用身份和授权流程，不能证明所有第三方都可用或全部不可用。后续按真实目标 SDK、应用身份与授权结果判断，不预先排除私有接口，也不猜测受保护端点的请求和签名格式。

## 原生工具链与结果边界

本机官方工具下载页面实际跳转华为账号登录；依据 [SDL 上游工作流](https://github.com/libsdl-org/SDL/blob/main/.github/actions/setup-harmonyos-toolchain/action.yml)尝试固定 6.1.1.280 镜像，本机及实际 Windows CI 均连接超时。还核对了两家 DoH 的真实 DNS 地址，并在保留 TLS 证书校验的情况下尝试直达，TCP 可建立但 TLS 握手超时，不能归因于 DNS 单一原因。云端首次运行 `41292708491` 的 SDK 任务明确失败在下载，未启动编译；日志已保留。随后将依赖外部 SDK 的编译单独设为显式手动入口，未添加 continue-on-error，也没有把下载失败改称编译成功。

没有代登、绕过账号校验或修改系统 SDK 配置。本机已验证构建脚本能写出 blocked 报告且 artifact 为 null。当前仍没有本轮 HAP、原生编译或模拟器运行证据。要执行原生验证入口，需可访问的 SDK 包及其预期哈希；实际目标仍是现有 API24 编译/兼容 API18 工程，不用低版本 OpenHarmony SDK 替代 HarmonyOS 结果。

原生构建命令（PowerShell 7）：

```text
pwsh -NoProfile -File SuiningWatch/scripts/build.ps1 -DevEcoHome "<DevEco Studio 根目录>"
pwsh -NoProfile -File SuiningWatch/scripts/build.ps1 -CommandLineHome "<command-line-tools 根目录>"
```

可选环境变量为 DEVECO_STUDIO_HOME、HARMONY_COMMANDLINE_HOME、DEVECO_SDK_HOME。二选一指定 IDE/CLI，避免混用。默认构建报告位于被 Git 忽略的 `SuiningWatch/artifacts/native-build.json`。安装模拟器仍需已运行的 Wearable 模拟器和 HDC，现有 `run-emulator.ps1` 对应 DevEco 布局；本轮未生成模拟器运行证据。

即使 SDK 编译通过，仍待真实设备确认：目标型号/系统安装兼容性、佩戴与心率可用性、计步器重启行为、权限系统弹窗、休眠与功耗、网络切换、通知可达性及私有服务访问资格。当前没有连续后台采样、GPS、血氧、ECG、睡眠或跌倒检测的新能力，不用模拟输入宣称这些功能完成。
