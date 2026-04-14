# go-blockchain

一个按课程大作业节奏逐步实现的 Go 区块链模拟仿真项目。

当前仓库的目标不是一次性把所有功能堆出来，而是按照阶段推进：

1. 先建立稳定的工程骨架
2. 再逐步实现区块链、交易、钱包、UTXO、共识、网络、GUI 等模块
3. 每个阶段同步补充可读、可讲解、可答辩的文档

## 当前进度

- Plan 1：项目初始化，已完成
- Plan 2：基础区块与区块链，已完成
- Plan 3：最简交易模型，已完成
- Plan 4：钱包系统，已完成
- Plan 5：UTXO 交易模型，已完成
- Plan 6：交易签名与验证，已完成
- Plan 7：UTXO 缓存与查询优化，已完成

## 仓库结构

```text
cmd/
  go-blockchain/        程序入口
internal/
  blockchain/           后续区块、区块链、交易主线
  cli/                  命令行入口
  config/               默认配置
  wallet/               后续钱包模块边界
  network/              后续网络模块边界
  gui/                  后续 GUI 模块边界
data/                   本地运行数据目录
docs/
  task.md               总体开发主流程
  plan/                 分阶段落地文档
tests/                  后续集成测试与验收脚本目录
```

## 当前可运行能力

当前仓库已完成初始化阶段的最小 CLI 骨架，支持：

- `--help`
- `version`
- `about`
- `doctor`

示例：

```bash
go run ./cmd/go-blockchain --help
go run ./cmd/go-blockchain version
go run ./cmd/go-blockchain doctor
```

## 文档怎么读

### 1. `docs/task.md`

这是整个大作业的总流程文档，用来说明完整开发路线和阶段顺序。

### 2. `docs/plan/README.md`

这是分阶段文档体系说明，介绍每个计划文档应该怎么写、怎么维护。

### 3. `docs/plan/plan1.md`

解释项目初始化阶段为什么这样做、目录怎么设计、CLI 和配置层怎么落地。

### 4. `docs/plan/plan2.md`

解释基础区块链阶段应该实现什么、代码会落在哪些位置、关键数据结构和开发顺序是什么。

### 5. `docs/plan/plan3.md`

解释最简交易模型如何落地，包括交易结构、coinbase 原型交易、区块内交易存储、CLI 演示方式和当前简化边界。

### 6. `docs/plan/plan4.md`

解释钱包系统如何落地，包括密钥生成、地址生成、钱包持久化、多钱包管理和 CLI 使用方式。

### 7. `docs/plan/plan5.md`

解释 UTXO 交易模型如何落地，包括输入引用历史输出、找零逻辑、UTXO 扫描、余额计算方式和 CLI 演示流程。

### 8. `docs/plan/plan6.md`

解释交易签名与验证如何落地，包括签名副本、引用输出校验、签名验证流程以及 CLI 演示方式。

### 9. `docs/plan/plan7.md`

解释 UTXO 缓存与查询优化如何落地，包括缓存键设计、重建索引、增量更新和 CLI 演示方式。

## 开发原则

1. 每个阶段先明确边界，再落地实现。
2. 每次开发时同步更新 `docs/plan/`，不是最后补文档。
3. 小更改直接提交并 push。
4. 大更改走 PR，并完成审核后合并。
5. 先打通命令行主链路，再扩展 GUI 和更复杂功能。

## 下一步

当前将进入 Plan 8：Merkle 树。

这个阶段会开始建立：

1. 交易哈希聚合
2. Merkle 树构建
3. MerkleRoot 写入区块头
4. 区块完整性验证增强

后续实现时会继续同步补充代码和文档。
