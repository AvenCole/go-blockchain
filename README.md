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
- Plan 8：Merkle 树，已完成
- Plan 9：工作量证明共识，已完成
- Plan 10：交易池与经济模型，已完成
- Plan 11：网络模拟，已完成
- Plan 12：安全校验与攻击模拟，已完成
- Plan 13：GUI 演示层，已完成
- Plan 14：性能实验与优化，已完成
- Plan 15：实验报告与答辩材料整理，已完成
- Plan 16：Script / OP 虚拟机，已完成
- Plan 17：脚本可视化与 GUI 演示增强，已完成
- Plan 18：分叉处理与最长链切换，已完成
- Plan 19：链重组后的 Mempool 恢复策略，已完成
- Plan 20：GUI 网络运维台与内置命令控制台，已完成
- Plan 21：链切换 / 重组状态可视化，已完成
- Plan 22：最近链事件面板，已完成

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

当前仓库当前已支持的核心 CLI 能力包括：

- `createwallet`
- `listaddresses`
- `createblockchain`
- `send`
- `printmempool`
- `mine`
- `getbalance`
- `printchain`
- `showscript`
- `simdouble`
- `runperf`
- `startnode`
- `simfork`
- `simreorg`
- `showreorg`
- `showevents`

示例：

```bash
go run ./cmd/go-blockchain createwallet
go run ./cmd/go-blockchain showscript <address>
go run ./cmd/go-blockchain createblockchain <miner-address>
go run ./cmd/go-blockchain printchain
```

GUI 默认使用独立数据目录：

- CLI / 节点：`./data`
- GUI：`./data/gui-desktop`

如需修改 GUI 数据目录，可设置环境变量：

- `GO_BLOCKCHAIN_GUI_DATA_DIR`

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

### 10. `docs/plan/plan8.md`

解释 Merkle 树如何落地，包括交易哈希叶子、奇数节点补齐、MerkleRoot 写入区块头以及完整性校验方式。

### 11. `docs/plan/plan9.md`

解释工作量证明如何落地，包括 Nonce、Difficulty、目标值判断、挖矿循环和区块校验方式。

### 12. `docs/plan/plan10.md`

解释交易池与经济模型如何落地，包括 Mempool、手续费、coinbase 奖励、显式挖矿流程和 CLI 演示方式。

### 13. `docs/plan/plan11.md`

解释网络模拟如何落地，包括节点启动、节点发现、交易广播、区块广播和基础同步流程。

### 14. `docs/plan/plan12.md`

解释安全校验与攻击模拟如何落地，包括双花检测、非法交易拦截、非法区块拒绝和 CLI 攻击演示入口。

### 15. `docs/plan/plan13.md`

解释 GUI 演示层如何落地，包括 Wails 桌面壳、React 19 + TypeScript + MUI 前端、多页面结构和真实后端绑定。

### 16. `docs/plan/plan14.md`

解释性能实验与优化如何落地，包括 cache vs scan 对比、CLI 实验入口和结果输出文件。

### 17. `docs/perf/`

保存性能实验输出文件，便于答辩和实验报告直接引用。

### 18. `docs/plan/plan15.md`

解释实验报告、架构说明、答辩脚本和检查清单为什么要单独整理成正式交付物。

### 19. `docs/plan/plan16.md`

解释 Script / OP 虚拟机如何落地，包括最小 P2PKH 指令集、脚本执行流程、兼容旧交易的原因和 CLI 演示方式。

### 20. `docs/plan/plan17.md`

解释如何把脚本系统在 GUI 中可视化展示，并同步处理 GUI 数据库句柄的稳定性问题。

### 21. `docs/plan/plan18.md`

解释如何保存侧链块、在更长分叉出现时切换到最长链，并通过 CLI 演示这一过程。

### 22. `docs/plan/plan19.md`

解释链重组后如何恢复旧主链掉下来的合法交易，并同步清理已在新主链确认的池中交易。

### 23. `docs/plan/plan20.md`

解释如何把网络节点操作和 CLI 命令执行能力整合进 GUI 工作台。

### 24. `docs/plan/plan21.md`

解释如何持久化最近一次链重组状态，并在 CLI 与 GUI 中做可视化展示。

### 25. `docs/plan/plan22.md`

解释如何把最近链事件列表持久化，并在 CLI 与 GUI 中展示最近事件序列。

## 开发原则

1. 每个阶段先明确边界，再落地实现。
2. 每次开发时同步更新 `docs/plan/`，不是最后补文档。
3. 小更改直接提交并 push。
4. 大更改走 PR，并完成审核后合并。
5. 先打通命令行主链路，再扩展 GUI 和更复杂功能。

## 下一步

当前已完成到 Plan 22：最近链事件面板。

当前新增的关键能力包括：

1. `scriptSig` / `scriptPubKey`
2. 最小 P2PKH Script VM
3. `showscript` CLI 演示入口
4. 新交易走脚本执行验证，同时兼容旧交易
5. GUI 钱包/区块页面可直接展示脚本信息
6. 更长分叉出现时可自动切换到最长链
7. 链重组后会恢复仍然合法的旧主链交易到 Mempool
8. GUI 已支持网络节点管理和内置 CLI 控制台
9. CLI / GUI 已支持最近一次链重组状态展示
10. CLI / GUI 已支持最近链事件列表展示

后续如果继续扩展，优先方向会是：

1. 更复杂的脚本类型
2. 更细粒度的链事件类型
3. 更接近 Bitcoin Core 的多窗口信息组织
4. 更接近真实节点的 orphan / reorg 广播策略
