# 华为运动健康参考材料：取舍与整合

基线：`preview / 9de7e069b69de799dd43d584e1434b4b2ff15050`。输入为用户提供的 `华为运动健康_分析产物_20261101.zip`，内含报告、Manifest、九份类签名、字符串池和复现脚本，未包含 APK。README 列出的 19 个文件 SHA-256 全部吻合。本次只读取产物，没有执行包内脚本或请求其中的私有地址。

## 结论与实现

**采纳按来源读取最新采样并用于首页展示；复用已有采样、日统计和增量同步体系。暂不引入 Android/HMS SDK、私有 Binder/REST、离线任务表或新告警算法。** 这直接补足当前统一采样写入与家属首屏之间的读取入口，不改变 Go/Vue/Tauri 与 MySQL 主库、PG 分析副库架构。

| 参考设计 | 当前项目与取舍 |
| --- | --- |
| `DataController.readLatestData(List<DataType>)` | 类签名产物 `normalized_result.methods[9]` 确实存在。新增通用 `POST /api/v1/health/latest`，家属首页通过 `LatestHealthPanel` 消费，不绑定华为实现 |
| `hihealth_dataclient` 来源归因 | 字符串池 `[301235]` 有 user/device/app/data_source_id。复用 `Sample.source/sourceDevice/sourceVersion` 与 `Record.oldId/deviceId/bindingVersion`，各来源分别返回，不自动混合或累加 |
| 原始点、区间、日聚合与同步锚点 | `sample_sequence` `[301250]`、`sync_anchor` `[301255]` 和日统计索引 `[1329]` 支持这种分层思路；项目已有 `health_sample`、`metric_daily`、`health_change` 与 revision 游标，不重复建表，不给所有业务表新增 sync_status |
| 按设备能力开放功能 | `DeviceManager.queryDeviceCapability` 签名存在；保留项目已有指令 profile 与不支持动作拒绝机制，不凭型号或类名宣称新传感器能力 |
| Health Connect / Health Kit / Wear Engine 接入 | 当前没有 Android 手机宿主、授权配置及真实供数验证。SDK 方法存在不能证明已获第三方访问权；不添加空壳连接器或把现有数据改标为华为数据 |

## 对原报告的关键纠正

- Manifest 第 327–329 行只声明写步数、活动热量、睡眠，第 5427–5431 行声明 SDK service。它们不能证明实际同步，更不能证明华为心率与血氧已进入 Health Connect。官方要求可用环境和逐类授权；Health Connect 支持 Android 9+，14+ 内置系统。[Android 可用性](https://developer.android.com/health-and-fitness/health-connect/availability)、[权限与数据类型](https://developer.android.com/health-and-fitness/health-connect/data-types)。
- `offline_data_task` 的 DDL `[50745]` 含 country/region/city/political/downloadProgress 等字段，指向地理或离线下载内容，不能当成健康补采任务的设计证据。
- 原报告“没有 FallDetection 痕迹”与自身字符串池矛盾：有 50 条 `FallDetection`、3 条 `fall_detection`，例如 `[57611] DEVMGR_FallDetectionManager`、`[70260] FallDetectionEvent`。这些线索既不能证明有公开第三方检测接口，也不能推导出必须自研算法。
- `MonitorItem` 实际为 10 个业务常量，另外两个字段为 CREATOR/mName。心率告警状态不能当心率数值；`sampleEvent` 路径常量 `[29583]` 也不能证明请求方向、鉴权、回调格式或开放资格。本次不移植相关告警、通知或私有接口。
- 本地推理库存在不代表全部算法都运行在手机，也不能排除固件、云端参与。

## 新接口合同

`POST /api/v1/health/latest` 使用现有会话及 CSRF，管理员传 `{ "oldId": 1 }`；家属可传 `{}` 使用会话关联老人，并实时核对当前家属归属。沿用既有岗位与归属检查，不增加新的授权模型。

成功响应为 `{ "returnCode": "1111", "oldId": 1, "items": [...] }`，每项复用 `health.Record`，保留 sampleId、deviceId、bindingVersion、revision 和完整 Sample。

- 每个精确 `(source, sourceDevice, metric, semantics)` 返回一条未删除的存量采样。MySQL 使用字节身份分组，避免大小写或尾随空格混组。
- 按 `Sample.start` 的精确时间降序选取；完全同刻按 revision 降序。SQL 微秒索引的并列候选全部进入比较，不先截成一行。
- 区间展示开始和结束时间；日累计明确标成快照，不当作今日汇总。前端不跨来源相加；累计类型跨来源相加可能重复计数，也与 [Health Connect 读取说明](https://developer.android.com/health-and-fitness/health-connect/read-data) 提醒一致。
- 订正读取当前版本；删除后可回退到其他未删除记录，但保留其原始时间与归属。旧绑定历史不会改成当前绑定采样。`legacy-unzoned` 明确显示原时间与时区未确认。
- 无数据返回空数组；首页给出空状态和已有日常记录入口。读取失败可以刷新重试；复用已有取消、超时、卸载处理。它是存量读取，不声称设备实时在线或已接入华为云。
- 最多处理 10000 个最新索引时间候选，超出返回 413 并指向已有增量接口，不静默截断。无数据库迁移或健康写入规则修改。

## 验证与交付

新增接口前，两条行为用例真实访问路由得到 404；实现后，MySQL 与 PG 均验证了来源/语义隔离、纳秒级同微秒选择、同刻 revision 决胜、订正、删除回退、归属复核和 CSRF。组件三条用例验证多来源、零值、区间、空状态、失败重试及卸载取消。该组件测试使用 HTTP 夹具，不是数据库浏览器联调。

新组件纳入现有 vue-tsc 范围；首次暴露 aria-busy 的字符串类型错误，改用布尔绑定后通过。受影响包 vet、路由权限测试与生产构建通过；既存大 chunk 提示保留。OpenAPI 显式登记请求、响应与 413，并与 108 条当前路由一致。

原始日志和页面截图放在外层 outputs，最终 SHA、CI、冒烟、PR 与合并结果写入工作记录及 PR。没有运行无关本地整套测试、耐久基准、真实设备或原生打包。回退可撤销本批提交，无数据迁移或数据删除。
