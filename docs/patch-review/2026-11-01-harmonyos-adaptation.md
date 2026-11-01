# HarmonyOS 手表接入适配（2026-11-01）

## 结构与通信方式

现有系统由 Vue 管理/家属网页、Go HTTP API、`SuiningWatch` ArkTS Stage 穿戴应用组成。原有手表端具备密钥认证、绑定码、心跳、离线队列和 SOS，但健康采样来自随机模拟。

本轮选择原生 HarmonyOS 应用直接通过 HTTP(S) 访问 Go，使用系统 SensorServiceKit、AbilityKit、NetworkKit。无需 Android 手机中转、蓝牙配对或华为账号 OAuth；“连接”指网络认证，“绑定授权”指设备与长者的后台关联，传感器授权另由手表系统确认。未接入华为运动健康历史数据，也未将静态 APK 中的私有接口当成可用服务合同。

用户允许使用私有接口/SDK，但本链路可用系统公开 Kit 实现，不额外引入不稳定的私有端点或依赖。支持的是可安装本 ArkTS 应用、提供相应传感器和权限的 HarmonyOS 穿戴设备；不据此宣称所有华为旧表/轻量系统均兼容。

## 改动位置与作用

| 位置 | 作用 |
| --- | --- |
| `SuiningWatch/entry/src/main/ets/services/HealthSensors.ets` | 枚举原生传感器，按硬件申请 READ_HEALTH_DATA / ACTIVITY_MOTION；一次性心率/计步读取，10 秒超时，退出时撤销监听，拒绝授权不生成假数据 |
| `SuiningWatch/entry/src/main/ets/model/NativeSampling.ets` | 心率 instant、原始累计计步转 interval 增量；首次建立基线，绑定/时钟/计数回退重置；保留 native timestamp 与应用接收时间的区别 |
| `SuiningWatch/entry/src/main/ets/model/Protocol.ets` | 登录协议版本 1 与最多 200 条批量协商、应用版本、原生来源元数据；真机地址要求 HTTPS，HTTP 仅限回环联调 |
| `SuiningWatch/entry/src/main/ets/services/WatchApi.ets` | 登录声明 harmonyos 平台；检验后端能力，禁用 HTTP 缓存，保留超时、密钥鉴权、句柄销毁和单次写入语义 |
| `SuiningWatch/entry/src/main/ets/services/WatchStore.ets` | 批次持久化失败后阻止进一步采样/覆盖，重启后恢复；不在状态不确定时丢弃原队列 |
| `SuiningWatch/entry/src/main/ets/pages/Index.ets` | 真实采样入口、权限提示、原生数据显示、模拟调试入口；绑定变更清除基线；派生数据保存成功后才推进计步基线 |
| `SuiningWatch/entry/src/main/module.json5`、`resources/base/element/string.json` | 声明传感器权限、用途文字与前台使用场景 |
| `SuiningWatch/entry/src/main/resources/base/profile/network_config.json` | 默认拒绝明文流量，回环模拟器地址例外；不关闭 TLS 证书验证 |
| `SuiningWatch/AppScope/app.json5`、`entry/oh-package.json5`、`package*.json` | 手表版本 0.2.0、便携验证入口与固定工具依赖 |
| `智慧养老手表管理系统/backend/internal/handlers/handler_device_protocol.go` | 在原 device/login 内增加可选平台/版本校验与能力响应；继续使用现有 secret、停用状态、bindingVersion，不增加平行账号体系或数据库表 |
| `智慧养老手表管理系统/frontend/src/components/HarmonyWatchSetup.vue`、`src/views/device.vue` | 登记完成/生成绑定码时展示真机与模拟器接入步骤，校验并复制后台地址；不把密钥放入 URL |
| `智慧养老手表管理系统/frontend/src/api/health.ts`、`src/components/LatestHealthPanel.vue` | 展示原生来源、应用接收/观测区间；继续按来源和语义分别显示，区间步数不显示成“今日总步数” |
| `scripts/generate-openapi.py`、后端 `internal/apidoc/openapi.json` | 补充 device/login 的增量协议合同；现有路由总数仍为 108 |
| 手表 `tests/harmony.test.cjs`、`tests/load-ets.cjs`、`tests/backend.e2e.cjs`；后端 `harmony_watch_test.go`；前端组件测试 | 验证转换、权限/取消、保存失败、传输与真实后端往返、来源展示与接入地址校验 |

## 连接、授权、数据流

1. 管理员在“设备管理”登记设备，保存仅显示一次的设备密钥。手表设置填入可访问的 HTTPS 主机、设备号与密钥；密钥只驻留内存，重启后重新输入。
2. `POST /api/v1/device/login` 发送 `{deviceNo, secret, platform:"harmonyos", protocolVersion:1}`。成功返回 `returnCode:"1111"`、`protocolVersion:1`、`maxBatchSize:200` 及原有绑定信息。平台/版本不支持返回 400；凭据错误沿用 HTTP 200 失败包络，必须检查 returnCode。旧客户端不传 platform 时合同不变。
3. 后台为长者生成设备绑定码；手表调用 `device/bind`，随后重新登录取得 oldId 和 bindingVersion。家属邀请码不能代替设备绑定码。
4. 首页主动申请系统权限并读取支持的传感器。心率有效值上传瞬时值；计步器累计值首读只建基线，后读上传相邻观测间的差。每条记录带 `source=suining-harmonyos`、`sourceDevice=设备号`、`simulated=false`。模拟来源保持 `suining-watch-sim`。
5. 传感器 timestamp 保留为元数据，不转换成 UTC。记录 start/end 使用应用接收时间并标注 `timestampBasis=app-received`；因系统调度造成的误差没有被包装成精准硬件采样时间。
6. “记录”页上传 `device/healthBatch`，每次先刷新绑定；后端事务复核凭据、设备状态、绑定版本与来源设备，按稳定 recordId/version 幂等。队列不跨后台、设备、长者或绑定版本混传。重试复用原记录，重绑后旧记录留待核查。
7. `device/heartbeat` 更新在线状态；家属会话经 `health/latest` 读取当前关联长者，网页显示来源与原生观测区间。SOS 沿用现有绑定版本、二次确认与 requestId 幂等。

## 验证结果与部署边界

- 手表便携检查 16 项通过；SDK 传感器/权限边界使用测试替身，不是物理传感器验收。
- Vue 类型检查、50 项测试、生产构建通过。构建仍有既有大块体积提示。
- 新增后端协议/原生数据回归在本机隔离 MySQL、PostgreSQL 均通过，handler vet 通过。
- 独立空库 Go 服务 HTTP 联调 8 项通过，使用生产 WatchApi/NativeSampling，验证认证、绑定、心跳、两种来源上传及重试、最近样本读取、SOS 和重绑隔离。传感器输入为已知测试数值，没有真表数据。
- 已运行原生构建脚本，但机器缺少 DevEco Node/SDK，构建未启动；没有生成本轮 HAP、签名、安装或真表权限/传感器验证。Node 转译不能证明 ArkTS 编译通过。

部署时先更新所有后端实例，再更新手表客户端。HTTPS 使用受设备信任的证书，后台需能从手表网络访问；模拟器继续使用现有 HDC 回环转发。需在装有 DevEco 的环境执行 `pwsh -File SuiningWatch/scripts/build.ps1 -DevEcoHome <安装目录>`，完成签名安装、权限允许/拒绝、连续两次计步、离线重试与重绑验证后，才能认定真表适配验收完成。当前仍为用户主动前台采集与上传，没有持续后台任务；GPS、睡眠、厂商云健康数据导入不在本轮实现范围。

## SDK 合同依据

- [华为 Sensor 开发概述](https://developer.huawei.com/consumer/en/doc/harmonyos-guides-V5/sensor-overview-V5)
- [OpenHarmony Sensor API 声明](https://github.com/openharmony/interface_sdk-js/blob/master/api/@ohos.sensor.d.ts)：getSensorList、once/off、HeartRateResponse、PedometerResponse。
- [AbilityAccessCtrl API 声明](https://github.com/openharmony/interface_sdk-js/blob/master/api/@ohos.abilityAccessCtrl.d.ts)：requestPermissionsFromUser。
- [HTTP 与网络安全配置](https://github.com/openharmony/docs/blob/master/en/application-dev/network/http-request.md)：NetworkKit、网络配置与 TLS 行为。

引用的是平台 API 合同，不代表目标手表已具备对应硬件、已获权限或本轮原生构建已通过。
