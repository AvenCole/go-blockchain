# 答辩演示脚本

## 一、总演示思路

答辩时建议按下面顺序演示：

1. 先讲系统目标与模块
2. 再演示 CLI 主链路
3. 再演示 GUI
4. 再演示网络与安全
5. 最后展示性能实验结果

## 二、开场介绍

可以这样开场：

“本项目是一个基于 Go 的区块链模拟仿真系统，采用分阶段实现方式，已经完成从区块、交易、钱包、UTXO、签名、Script / OP 虚拟机、Merkle、PoW、Mempool、网络模拟，到 GUI 和性能实验的完整链路。” 

## 三、建议演示顺序

### 1. 钱包创建

命令：

```bash
go run ./cmd/go-blockchain createwallet
go run ./cmd/go-blockchain createwallet
go run ./cmd/go-blockchain listaddresses
```

讲解重点：

1. 已经有真实钱包和地址
2. 地址不是手写字符串
3. 钱包支持持久化

### 2. 初始化区块链

命令：

```bash
go run ./cmd/go-blockchain createblockchain <miner-address>
go run ./cmd/go-blockchain printchain
```

讲解重点：

1. 创世区块已建立
2. 区块头已经包含 MerkleRoot、Difficulty、Nonce
3. PoW 校验已经成立

### 2.5 脚本系统展示

命令：

```bash
go run ./cmd/go-blockchain showscript <miner-address>
```

讲解重点：

1. 地址可以映射成标准 P2PKH 锁定脚本
2. 当前系统已经支持 `OP_DUP`、`OP_HASH160`、`OP_EQUALVERIFY`、`OP_CHECKSIG`
3. 交易验证已经升级为脚本执行，而不只是代码里直接写死的签名判断

### 3. 发送交易 + Mempool + 挖矿

命令：

```bash
go run ./cmd/go-blockchain send <miner-address> <alice-address> 20 2
go run ./cmd/go-blockchain printmempool
go run ./cmd/go-blockchain mine <miner-address>
go run ./cmd/go-blockchain getbalance <miner-address>
go run ./cmd/go-blockchain getbalance <alice-address>
go run ./cmd/go-blockchain printchain
```

讲解重点：

1. 交易先进入 Mempool
2. 挖矿后新区块生成
3. coinbase 奖励和手续费都进入矿工收益
4. 输入输出是 UTXO 风格
5. 交易有签名校验
6. `printchain` 中可以直接看到 `scriptSig` 和 `scriptPubKey`

### 4. 双花攻击模拟

命令：

```bash
go run ./cmd/go-blockchain simdouble <from> <to1> <to2> 20
```

讲解重点：

1. 第一笔交易进入池中
2. 第二笔冲突交易被系统拒绝
3. 双花检测不是人工判断，而是系统自动生效

### 5. 网络模拟

开两个终端：

```bash
go run ./cmd/go-blockchain startnode 127.0.0.1:3010
go run ./cmd/go-blockchain startnode 127.0.0.1:3011 127.0.0.1:3010
```

讲解重点：

1. 节点之间可以发现彼此
2. 区块和交易能够传播
3. 新节点能同步已有链

### 5.5 最长链切换演示

命令：

```bash
go run ./cmd/go-blockchain simfork <miner-address> 2
```

讲解重点：

1. 系统允许保存侧链块
2. 当侧链长度超过当前主链后，tip 会切换
3. 切换后会重建 UTXO，保证账本状态一致

### 5.6 链重组后交易恢复演示

命令：

```bash
go run ./cmd/go-blockchain simreorg <miner-address> <receiver-address> 20 1
```

讲解重点：

1. 交易先在旧主链中被确认
2. 更长分叉切换后，这笔交易会掉出主链
3. 如果它在新主链下仍然合法，系统会把它恢复回 Mempool

### 5.7 最近一次重组状态展示

命令：

```bash
go run ./cmd/go-blockchain showreorg
```

讲解重点：

1. 系统会记录最近一次重组时间
2. 可以展示旧高度 / 新高度
3. 可以展示恢复交易数和清理交易数

### 5.8 最近链事件列表展示

命令：

```bash
go run ./cmd/go-blockchain showevents 5
```

讲解重点：

1. 最近事件不是 GUI 假数据，而是链元数据真实记录
2. 可以回顾最近若干次链变化
3. 便于解释系统最近发生过哪些 reorg / 链切换

### 5.9 孤儿块缓冲说明

讲解重点：

1. 如果节点先收到子块、后收到父块，系统不会直接把子块当坏块丢掉
2. 节点会先缓存孤儿块，并请求缺失父链
3. 当父块到达后，孤儿块会被自动重试导入
4. GUI 网络页还能看到当前 orphan 数量

### 5.10 更细粒度的最近链事件

讲解重点：

1. 最近事件不只包含 reorg
2. 还包含 genesis、main_block、fork_block
3. 这样老师能看到系统正常增长、分叉暂存和重组切换的完整链行为

### 5.12 第二种脚本模板 P2PK

建议命令：

```bash
go run ./cmd/go-blockchain showscript <address> p2pk
go run ./cmd/go-blockchain sendp2pk <from> <to> 20 1
```

讲解重点：

1. Script VM 不再只有 P2PKH 一种模板
2. 现在还能直接演示 P2PK
3. 说明脚本系统具备模板扩展能力

### 5.13 教学型多重签名

建议命令：

```bash
go run ./cmd/go-blockchain sendmultisig <from> 2 <addr1,addr2> 20 1
go run ./cmd/go-blockchain spendmultisig <addr1,addr2> <source-txid> 0 <to> 10 1
```

讲解重点：

1. Script VM 现在不只支持单签模板
2. 还能演示多个签名者共同解锁同一输出
3. 这说明脚本系统不只是“换个壳”，而是真能表达不同花费条件

### 6.5 GUI 脚本交易模板切换

建议在 GUI 交易页讲解：

1. 可以切换 P2PKH / P2PK / MultiSig
2. 前端只是选择模板和参数
3. 真实脚本交易仍然由 Go 后端构造

### 6.6 GUI 多签花费向导

建议在 GUI 交易页讲解：

1. 可以看到当前未花费多签输出列表
2. 可以直接选择一个多签输出进行花费
3. 签名者 CSV、目标地址、金额和手续费仍然走真实后端构造

### 6.7 主工作台总览

讲解重点：

1. Dashboard 现在不是单一摘要，而是多区域总控台
2. 在同一页可以同时看到链状态、最新区块、钱包、多签输出、Mempool 和节点活动摘要
3. 这更接近桌面客户端 / Bitcoin Core 的信息组织方式

### 5.11 节点最近网络事件

讲解重点：

1. 网络页现在不仅显示高度和 peer
2. 还可以看到节点最近做了什么，例如 peer 加入、孤儿块缓存、父块补取、孤儿块恢复
3. 这让网络层行为不再只是口头说明

### 5.12 最新高度通告

讲解重点：

1. 节点在自己挖出新区块后会主动向其他 peer 通告最新高度
2. 节点在导入新区块推进 tip 后也会主动通告
3. 这样网络同步不再只是被动请求，还有主动提示行为

### 5.13 GUI 节点链控制

讲解重点：

1. GUI 网络页现在不只是看节点状态，还能直接初始化某个节点的本地链
2. 可以指定一个节点直接发交易、直接挖矿
3. GUI 控制台也支持 `nodeinit`、`nodesend`、`nodemine`，适合同时演示图形界面和终端链路

### 6. GUI 演示

直接打开桌面程序：

```text
build/bin/go-blockchain-gui.exe
```

建议在 GUI 中演示：

1. Dashboard
2. 钱包页面
3. 区块浏览
4. 交易排队
5. 挖矿操作
6. 钱包页和区块页中的脚本展示
7. 网络页面与控制台页面

讲解重点：

1. GUI 不是静态壳
2. 页面调用的是真实 Go 后端
3. CLI 与 GUI 展示的是同一套能力

### 7. 性能实验

命令：

```bash
go run ./cmd/go-blockchain runperf 20
```

再展示：

1. `docs/perf/latest.md`
2. `docs/perf/latest.json`

讲解重点：

1. 全链扫描 vs UTXO 缓存
2. 为什么 UTXO 缓存更快
3. 系统不只是能跑，还有实验数据支撑

## 四、老师可能会问的问题

### Q1：你这个系统和真正比特币的差距在哪？

建议回答：

1. 当前已经覆盖区块链主链路的大部分核心机制
2. 当前已经实现教学型最小 Script VM，但还不是完整 Bitcoin Script 全指令集
3. 网络层也还是教学型本地模拟，不是完整比特币 P2P 协议

### Q2：为什么要分这么多阶段做？

建议回答：

1. 这样每一阶段目标清晰
2. 更容易验证
3. 更适合答辩讲解
4. 也能避免一开始把所有复杂性混在一起

### Q3：你系统里最有价值的部分是什么？

建议回答：

1. UTXO + 签名交易链路
2. Script VM 驱动的交易验证
3. Merkle + PoW 区块头
4. Mempool + 手续费 + coinbase 奖励
5. 网络模拟
6. GUI 和性能实验

## 五、答辩结束总结

最后可以这样收尾：

“这个系统已经完成了一个教学型区块链模拟系统从底层数据结构、交易安全、脚本验证、共识机制，到网络模拟、GUI 展示和性能实验的完整闭环。后续如果继续扩展，最优先的方向是更完整的脚本类型、更完整的分叉处理和更强的网络协议模拟。” 
