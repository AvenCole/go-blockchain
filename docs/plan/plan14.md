# Plan 14：性能实验与优化

## 一、这一阶段要解决什么问题

到 Plan 13 为止，系统已经具备完整的教学型链路，但如果没有实验数据，就很难在最终报告里证明优化确实有效。

Plan 14 的目标是：

1. 对比全链扫描与 UTXO 缓存查询
2. 输出可复现实验数据
3. 把实验结果写到文档文件

## 二、为什么这一阶段必须做

因为你后面最终要做实验报告和答辩。  
如果没有真实测量数据，性能优化就只能停留在口头描述。

## 三、本阶段范围

### 要做的内容

1. 提供全链扫描余额查询函数
2. 提供 UTXO 缓存余额查询对比
3. 提供 CLI `runperf`
4. 把结果输出到 `docs/perf/latest.json` 和 `docs/perf/latest.md`
5. 同步文档

### 这一阶段不做的内容

1. 不做复杂 profiling 平台集成
2. 不做数据库层深度微调
3. 不做网络性能实验

## 四、核心理论

### 1. 为什么对比“全链扫描 vs UTXO 缓存”

因为它们正好对应这个项目最关键的一组优化前后状态：

1. 未优化：每次余额查询都遍历整条链
2. 已优化：直接查 UTXO 缓存

### 2. 为什么实验输出要写成文件

因为这样：

1. 答辩时更容易展示
2. 后续写报告时更方便引用
3. 结果可复查

## 五、代码会落在哪些位置

1. `internal/blockchain/performance.go`
2. `internal/blockchain/performance_test.go`
3. `internal/cli/app.go`
4. `docs/perf/`
5. `docs/plan/plan14.md`

## 六、验收标准

1. 能运行 `runperf`
2. 生成 `docs/perf/latest.json`
3. 生成 `docs/perf/latest.md`
4. 能看到 cache vs scan 时间差异
5. 测试通过
6. 文档同步更新
