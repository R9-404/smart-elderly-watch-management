# 性能测量

历史本机 HTTP 延迟与 limiter 内存实验的原始记录已从当前分支清理，可从 Git 历史查阅。
本文不再引用已删除的逐请求数据或复述其性能数字。

当前性能基线应绑定源码 SHA、Go 与数据库版本、硬件和负载参数，并保留延迟分位数、
错误率及内存记录。现存 CI 功能记录不代表当前版本的性能或容量测量。

## JSON 字符串扫描：amd64 SSE2

`backend/internal/jsonscan` 为健康样本序列化、样本身份哈希及规则证据生成提供共享扫描器。
16 字节及以上在 amd64 使用手写 SSE2（`PCMPGTB`、`PCMPEQB`、`POR`、`PMOVMSKB`）；
短字符串使用 Go 查表。其他架构及 `-tags purego` 使用原逐字节判定。
控制字符、非 ASCII、引号、反斜杠及 `<>&` 仍返回 false，由原 JSON 编码器处理。
不改变编码字节、样本身份、密码学算法、数据库事务或业务判定。

汇编使用 Go ABI0、`NOSPLIT` 和不逃逸声明，不使用 cgo；仅当剩余长度不少于 16 时执行
向量加载，尾部逐字节读取，不依赖分配器填充或越界读取。amd64 使用基线 SSE2，无 AVX CPU 要求。

本机 Go 1.26.10 / Linux amd64 / Intel Xeon Platinum 8573C，`GOMAXPROCS=2`，
每项 150ms、三次结果中位数；环境存在调度波动。扫描基线为原 Go 字节条件，身份哈希基线为
同一测试二进制内保存的原实现，保持 SHA-256 路径相同；不以全局 purego 标签比较性能。

| 输入 | 原实现 ns/op | 新实现 ns/op | 说明 |
|---|---:|---:|---|
| 16 字节 ASCII 扫描 | 58.83 | 9.419 | 仅扫描器 |
| 64 字节 ASCII 扫描 | 144.8 | 17.09 | 仅扫描器 |
| 256 字节 ASCII 扫描 | 470.6 | 38.73 | 仅扫描器 |
| 1024 字节 ASCII 扫描 | 1852 | 164.7 | 仅扫描器 |
| 身份哈希：设备号及记录 ID 各 64 字节 | 1040 | 674.8 | 分配不变：128 B、2 次 |
| 身份哈希：设备号及记录 ID 各 256 字节 | 2658 | 1824 | 分配不变：1024 B、3 次 |

短字段未证明稳定的完整业务收益。32 字节身份字段第一次测量出现明显噪声，追加
`GOMAXPROCS=1`、500ms、三次对照的中位数为 613.7 → 634.6 ns/op，未显示加速。
以上结果不代表 HTTP 延迟、数据库吞吐或生产容量提升。

复现（在 backend 目录）：

```sh
go test ./internal/jsonscan ./internal/health ./internal/rules
go test -tags purego ./internal/jsonscan
go vet ./internal/jsonscan
GOMAXPROCS=2 go test ./internal/jsonscan -run '^$' -bench 'BenchmarkSafe/ascii' -benchtime=150ms -count=3
GOMAXPROCS=2 go test ./internal/health -run '^$' -bench BenchmarkIdentityScan -benchtime=150ms -count=3
```

已执行字节等价、向量/尾部边界、Linux guard page 测试及相关 health/rules 包测试；
Windows amd64 与 Linux arm64 仅完成交叉编译，未执行目标平台测试。未运行数据库压测。

## Linux amd64 原生内核

Linux amd64 在启用 CGO 的构建中使用 C/C++ 工具链。可在 backend 目录执行
`CGO_ENABLED=1 go build -trimpath -o suining ./cmd/server` 构建带原生内核的服务端。
`CGO_ENABLED=0`、`-tags purego` 及其他平台使用 Go 实现。

- 24/32 位无压缩 BMP 通过 Go 的尺寸和长度校验后，达到 4096 像素时由 C 批量转换 RGBA。
- 2048 条及以上的乱序睡眠记录在 C++ 中排序连续时间戳键，随后由 Go 原地重排样本。
  已按时间排序的数据直接进入原有聚合流程，带单调时钟的输入由 Go 排序。

Go 1.26.10、Linux amd64、Xeon Platinum 8573C 本机三次对照的中位数：

| 测量 | Go | 原生实现 |
|---|---:|---:|
| 1024×1024 24 位 BMP 解码，Go 基线直接写 RGBA 缓冲区 | 3.37 ms | 1.81 ms |
| 1024×1024 32 位 BMP 解码，Go 基线直接写 RGBA 缓冲区 | 4.77 ms | 3.17 ms |
| 10,000 条乱序睡眠记录完整聚合，含打包和重排 | 4.87 ms | 3.10 ms |
| 同一睡眠聚合的进程 CPU 时间 | 4.89 ms | 3.70 ms |

BMP 对照包含输出缓冲区分配。睡眠聚合的分配从 310,840 增至 474,680 B/op。
测量范围为解码和聚合函数，生产接口延迟仍需在真实负载下测量。
