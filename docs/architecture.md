# 系统架构说明

## 1. 项目目标

本项目是一个面向课程答辩的 Go 区块链模拟仿真系统。  
目标不是完全复制比特币全部实现，而是在教学范围内尽可能完整地覆盖区块链核心机制，并把这些机制做成：

1. 可运行
2. 可演示
3. 可解释
4. 可答辩

## 2. 整体模块结构

当前系统主要由以下模块组成：

```text
go-blockchain
├── cmd/go-blockchain          命令行入口
├── internal/blockchain        区块链核心
├── internal/wallet            钱包与地址
├── internal/network           网络模拟
├── internal/gui               GUI 后端服务层
├── docs/plan                  分阶段设计文档
├── docs/perf                  性能实验输出
└── frontend                   GUI 前端（React 19 + TS + MUI）
```

## 3. 核心链路

系统当前已经打通的主链路如下：

1. 钱包生成地址
2. 创世块创建
3. 交易进入 mempool
4. 矿工打包交易并创建 coinbase
5. 区块进行 PoW 挖矿
6. 区块写入链
7. UTXO 缓存同步更新
8. 网络节点广播区块和交易
9. GUI/CLI 同步读取真实链状态

## 4. 区块链核心模块

`internal/blockchain` 负责：

1. `Block`
2. `Blockchain`
3. `Transaction`
4. `Script / OP` 虚拟机
5. `UTXO` 查找与缓存
6. `MerkleTree`
7. `ProofOfWork`
8. `Mempool`
9. 区块与交易验证

### 4.1 Block

区块当前主要包含：

1. 时间戳
2. 交易列表
3. `MerkleRoot`
4. 前一区块哈希
5. 当前区块哈希
6. 高度
7. `Nonce`
8. `Difficulty`

### 4.2 Transaction

交易当前已经升级到：

1. UTXO 风格输入输出
2. 签名输入
3. 公钥证明
4. 手续费支持
5. coinbase 奖励交易
6. `scriptSig` / `scriptPubKey`
7. 最小 P2PKH Script VM

### 4.3 UTXO Cache

为了避免每次全链扫描，系统维护了单独的 UTXO 缓存。

缓存特点：

1. 使用 Pebble 存储
2. 通过 key 前缀区分链数据和 UTXO 数据
3. 支持增量更新
4. 支持全量重建

## 5. 钱包模块

`internal/wallet` 负责：

1. ECDSA 密钥对生成
2. 地址生成
3. Base58 编码
4. 地址校验
5. 多钱包集合持久化

当前地址方案是教学型、简化版比特币风格地址生成方案。

## 6. 网络模块

`internal/network` 负责本地 TCP 多节点模拟：

1. 节点启动
2. 版本握手
3. 节点发现
4. 区块同步
5. 交易广播
6. 区块广播

当前网络层是教学型模拟，不追求公网生产级 P2P 协议。

当前链管理层已经支持：

1. 保存侧链块
2. 在更长分叉出现时切换到最长链
3. 切换后重建 UTXO 状态
4. 恢复旧主链掉下来的合法交易到 Mempool
5. 持久化最近一次链重组状态
6. 持久化最近链事件列表

当前网络层还支持孤儿块缓冲：当区块乱序到达时，可先缓存子块并等待父块补齐。
当前 GUI 网络页还能展示节点最近的关键网络事件。

当前最近链事件类型已经覆盖：

1. genesis
2. main_block
3. fork_block
4. reorg

当前 Script VM 已经支持的模板包括：

1. P2PKH
2. P2PK
3. 教学型多重签名

GUI 交易页当前也已经可以直接发起多种脚本模板交易，而不再局限于普通 P2PKH 转账。

## 7. GUI 模块

GUI 使用：

1. Wails
2. React 19
3. TypeScript
4. MUI
5. React Compiler

### 7.1 GUI 后端层

`internal/gui` 负责把链、钱包、区块、交易池等能力包装成前端可直接使用的数据结构。

### 7.2 GUI 前端层

前端分为：

1. Dashboard
2. 钱包页面
3. 区块浏览
4. 交易与挖矿页面
5. 脚本相关信息展示
6. 节点网络操作
7. 内置 CLI 控制台

并且已经支持浅色/深色模式切换。

### 7.3 GUI 数据隔离

为了避免和 CLI / 节点模拟争抢同一个 Pebble 数据库句柄，GUI 默认使用独立数据目录：

1. CLI / 节点：`./data`
2. GUI：`./data/gui-desktop`

## 8. 安全边界

系统当前已经具备的安全边界包括：

1. 交易签名验证
2. 脚本执行验证
3. 非法输出引用拒绝
4. 双花拒绝
5. 非法区块导入拒绝
6. 网络侧 fail-closed 导入策略

## 9. 性能实验

系统当前已经支持：

1. 全链扫描余额查询
2. UTXO 缓存余额查询
3. `runperf` 性能实验命令
4. `docs/perf/latest.json`
5. `docs/perf/latest.md`

## 10. 当前系统定位

当前系统定位可以概括为：

**一个已经具备教学型完整链路的区块链仿真系统**

它并不是完整比特币实现，但已经覆盖了课程作业里最有价值的部分：

1. 区块
2. 交易
3. 钱包
4. UTXO
5. 签名
6. Script VM
7. Merkle
8. PoW
9. Mempool
10. 网络模拟
11. GUI
12. 性能实验

## 11. 后续可扩展方向

如果继续扩展，最合适的方向包括：

1. 更完整的分叉/最长链处理
2. 更复杂的脚本类型
3. 更复杂的网络协议
4. GUI 图形化链结构与脚本展示
5. 更完整的性能实验
