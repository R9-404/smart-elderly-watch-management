# 可选 PostgreSQL 分析库的故障排查与独立告警

本手册以 [`deploy/platform.compose.yml`](../../deploy/platform.compose.yml) 的当前配置为准。**MySQL 是唯一业务主库；PostgreSQL 仅承载异步分析投影**。PG 停机不能转化为业务写入成功的虚假报告，也不应导致 MySQL API 无法启动。

## 运行关系

- `app` 依赖 MySQL 健康；PG 的分析进程异步重试，不是应用启动门槛。
- `prometheus` 可以独立运行；`/healthz` 仅表示进程存在，真实业务就绪以 `GET /api/v1/service/status` 等专用状态为准。
- 当前仓库**已有** `deploy/alertmanager.yml` 和 Compose 可选 `paging` profile，不能再描述为“没有 Alertmanager 配置”。但默认启动**不会**自动启用 `paging`；存在配置文件不代表已联系真实值班人员。
- `deploy/prometheus.yml` 包含 Alertmanager 目标；未启用 `paging` 时相关连接可能失败，不意味着外部告警已成功送达。

## 受控环境下的验证

先使用独立测试数据库、临时密钥与无真实接收人的网关。确认 Docker、Compose 与部署配置通过：

```sh
docker compose --env-file deploy/platform.env -f deploy/platform.compose.yml config --quiet
docker compose --env-file deploy/platform.env -f deploy/platform.compose.yml up -d
```

在测试环境中停止 PG，观察主库业务仍正常，分析错误和积压有可见指标：

```sh
docker compose --env-file deploy/platform.env -f deploy/platform.compose.yml stop postgres
# 验证 GET /api/v1/service/status、照护写入、MySQL outbox 与 worker 状态
docker compose --env-file deploy/platform.env -f deploy/platform.compose.yml start postgres
# 核对断线期间事件被补偿、最终积压收敛
```

停止生产库、改变真实通知渠道或强制清空队列均不属于本文的自动测试步骤。故障演练应同时记录原始时刻、等待时间、恢复时间、队列容量及有无重复投影。

## 外部告警的真实启用条件

先由运维人员创建受控的 `deploy/secrets/oncall-webhook-url` 与 `deploy/secrets/oncall-token`，确认接收端地址为可验证的 HTTPS、凭据有效、人员具备独立应答流程；然后在隔离环境启用：

```sh
docker compose --env-file deploy/platform.env -f deploy/platform.compose.yml --profile paging up -d
```

核对 Prometheus 规则、Alertmanager 的健康/路由及真实受控接收端的接收记录，并验证断线、重复发送与恢复告警。独立 watchdog 与值班通道的故障域、凭据及顺序保证见 [滥用控制与独立监控](../reliability/abuse-and-monitoring.md)。

**三级状态必须区分：** 告警规则触发、接收端 HTTP 2xx 接收、值班人员实际接手。任何一级都不等于 SOS 已被成功处置。没有真正连接的值班端时，不得宣称已经部署了自动求救通知链。

## 恢复与升级边界

PG 失效后的重建只能基于 MySQL 的业务记录/outbox，不能将旧 PG 反向写成权威业务数据。按 [双库操作指南](../DUAL_DATABASES.md) 的 `analytics-backfill` 流程执行；要恢复受影响的 MySQL、图片和密钥，请使用 [灾难恢复手册](backup-restore.md)。
