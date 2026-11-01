# HarmonyOS 手表接入

手表工程位于 [SuiningWatch](../SuiningWatch/README.md)，网页在设备管理中提供登记、绑定码和接入引导，Go 后端负责设备认证、归属校验、健康采样入库和 SOS 工单。

## 通信流程

1. 管理员先创建长者档案，再登记设备，保存创建时展示的设备密钥。
2. 手表配置可访问的 HTTPS 后端地址、设备号和密钥，调用 `device/login`，携带 `platform=harmonyos`、`protocolVersion=1`。
3. 后端返回协议版本、批次上限、长者归属和 `bindingVersion`。输入设备绑定码调用 `device/bind` 后重新认证。
4. 手表申请心率和活动权限；心率按瞬时样本保存，累计计步先建立基线，再转换为区间增量。调试采样单独标记 `simulated`。
5. `device/healthBatch` 携带原绑定版本和稳定采样 ID。重试复用记录 ID，重新绑定后保留旧队列并阻止补传。
6. SOS 使用稳定 `requestId` 和原绑定版本。后台接收、通知回执、人工处置分别记录。

全部接口的响应与错误格式见 [接口合同](API_CONTRACTS.md)。设备密钥错误返回 401，设备停用返回 403，绑定变化返回 409；网页与手表展示后端消息和追踪编号。

## 开发与验收

- [构建、模拟器端口转发和联调命令](../SuiningWatch/README.md)
- [采样模块与授权实现](patch-review/2026-11-01-harmonyos-adaptation.md)
- [无真机阶段的生命周期与断网恢复检查](patch-review/2026-11-01-harmonyos-without-device.md)
- [华为运动健康参考资料分析](patch-review/2026-11-01-huawei-health-reference.md)

当前实现包含前台原生传感器读取与模拟采样；原生 SDK 编译、签名安装和实表传感器验收需要对应工具及设备。旧版本的原生构建记录保存在 `evidence/watch-import-20261026/`，其源码版本由记录中的 SHA 标识。

手表密钥驻留内存，采样队列保存在应用私有 Preferences。后台自动采样、GPS、睡眠阶段检测和本地队列加密尚未实现。
