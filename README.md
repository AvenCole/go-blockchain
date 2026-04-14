# go-blockchain

一个按课程大作业节奏逐步实现的 Go 区块链模拟仿真项目。

当前仓库的目标不是一次性把所有功能堆出来，而是按照阶段推进：

1. 先建立稳定的工程骨架
2. 再逐步实现区块链、交易、钱包、UTXO、共识、网络、GUI 等模块
3. 每个阶段同步补充可读、可讲解、可答辩的文档

## 当前进度

- Plan 1：项目初始化，已完成
- Plan 2：基础区块与区块链，已进入计划阶段

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

## 开发原则

1. 每个阶段先明确边界，再落地实现。
2. 每次开发时同步更新 `docs/plan/`，不是最后补文档。
3. 小更改直接提交并 push。
4. 大更改走 PR，并完成审核后合并。
5. 先打通命令行主链路，再扩展 GUI 和更复杂功能。

## 下一步

当前将进入 Plan 2：基础区块与区块链。

这个阶段会开始建立：

1. `Block`
2. `Blockchain`
3. 创世区块
4. 区块追加
5. 区块链遍历
6. 基础持久化

后续实现时会继续同步补充代码和文档。
