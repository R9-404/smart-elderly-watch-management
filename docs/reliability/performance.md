# 性能基线与可复现测量

按当前提交 SHA 与实际运行条件复现性能基线。旧微基准数据保留在 Git 历史。

## 当前可测量的关键路径

| 路径 | 代码/基准入口 | 注意事项 |
| --- | --- | --- |
| JSON 安全扫描 | `backend/internal/jsonscan` | amd64 SSE2 与 purego 路径须保持编码字节一致；包含转义字符和边界长度用例 |
| 健康采样身份 | `backend/internal/health/identity_scan_benchmark_test.go` | 不改变样本 ID、哈希及去重语义 |
| 健康统计热缓存 | `backend/internal/handlers/statistics_cache_benchmark_test.go` | 避免把缓存命中率等同于数据库写入速度；校验跨实例版本失效 |
| 睡眠窗口聚合 | `backend/internal/health/sleep_sort_native_test.go` | Linux amd64 可选原生排序路径；比较总时长、分配、结果与纯 Go 路径 |
| 限流窗口与列表分页 | `backend/internal/ratelimit/benchmark_test.go`、`backend/internal/handlers/list_pagination_benchmark_test.go` | 独立测试高基数、内存上界和错误率 |

部分 Linux amd64 加速路径依赖 CGO/C++ 编译器；不同架构、`-tags purego` 或 `CGO_ENABLED=0` 可能使用纯 Go。目标平台性能需在对应架构实测。

## 本地复现

在 `智慧养老手表管理系统/backend` 目录：

```sh
go test ./internal/jsonscan ./internal/health ./internal/rules
go test -tags purego ./internal/jsonscan
go vet ./internal/jsonscan
GOMAXPROCS=2 go test ./internal/jsonscan -run '^$' -bench 'BenchmarkSafe/ascii' -benchtime=500ms -count=3
GOMAXPROCS=2 go test ./internal/health -run '^$' -bench BenchmarkIdentityScan -benchtime=500ms -count=3
```

涉及数据库的性能评估必须在专用 MySQL/PG 测试实例上完成，并记录提交 SHA、Go 版本、CPU、CPU 限额、数据库版本/配置、数据规模、并发/请求混合、负载生成器、预热时间和重复次数。**报告 p50/p95/p99、吞吐、CPU、峰值 RSS、GC、网络错误、业务失败和重试量**，同时保存原始日志及构建来源。

对照改动前后的相同路径和负载，首先检查业务结果、事务、告警、幂等与权限没有退化，再评估性能。端到端 SOS 延迟与生产容量使用独立负载测试。

## 验收与保留

- 对 SIMD/CGO 优化使用字节/哈希等价、向量尾部、越界访问及 `-race` 检查；根据支持的平台执行纯 Go 回退。
- 如果提升很小或测试波动大，保留简单的 Go 实现，避免以额外复杂性换取不稳定微收益。
- 只引用与**同一代码提交**匹配的实验 artifact；被删除的历史性能日志仅可用于历史对比，不可冒充最新环境证据。
