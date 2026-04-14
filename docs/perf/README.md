# docs/perf 说明

这个目录用于保存性能实验的最新输出。

当前约定：

1. `latest.json`：结构化实验结果
2. `latest.md`：适合直接截图或复制到实验报告中的文本结果

推荐在答辩前重新运行一次：

```bash
go run ./cmd/go-blockchain runperf 20
```

这样可以得到最新的实验数据和展示材料。
