# Plan 10：交易池与经济模型

## 一、这一阶段要解决什么问题

前面的阶段已经让链具备：

1. 钱包
2. UTXO
3. 签名验证
4. UTXO 缓存
5. Merkle 树
6. PoW

但当前 `send` 还是“发一笔交易就立刻成块”，这和真实区块链流程差距还很大。

Plan 10 要解决的是：

1. 交易先进入 `Mempool`
2. 矿工从 `Mempool` 取交易打包
3. 区块有正式的 coinbase 奖励
4. 区块可以把手续费一并奖励给矿工

## 二、为什么交易池和经济模型必须一起做

因为在真实链路里：

1. 用户先发交易
2. 交易进入待打包池
3. 矿工挑选交易
4. 矿工打包新区块
5. 区块第一笔是 coinbase 奖励交易
6. 手续费最终归矿工

所以 `Mempool` 和奖励/手续费机制天然是一条链路。

## 三、本阶段范围

### 要做的内容

1. 定义 `Mempool`
2. 发送交易只入池，不直接出块
3. 增加显式 `mine` 命令
4. 矿工从池中取交易打包
5. coinbase 奖励 = 基础奖励 + 手续费
6. 提供 `printmempool`
7. 同步文档

### 这一阶段不做的内容

1. 不做网络广播的 mempool 同步
2. 不做复杂交易排序策略
3. 不做交易优先级市场
4. 不做动态区块大小限制

## 四、核心理论

### 1. Mempool 是什么

Mempool 可以理解成“待打包交易池”。

交易不是一发出就立刻写进区块，而是先进入池中等待矿工打包。

### 2. 为什么要有 coinbase 奖励

因为矿工需要激励。  
区块第一笔交易通常是系统生成的 coinbase 奖励交易。

### 3. 手续费从哪里来

在 UTXO 模型中：

1. 输入总额
2. 减去输出总额
3. 剩余部分

就是手续费。

## 五、代码会落在哪些位置

1. `internal/blockchain/blockchain.go`
2. `internal/blockchain/mempool.go`
3. `internal/blockchain/transaction.go`
4. `internal/blockchain/blockchain_test.go`
5. `internal/cli/app.go`
6. `internal/cli/app_test.go`
7. `docs/plan/plan10.md`

## 六、推荐设计

### 1. Mempool 存储

当前阶段可以直接把 mempool 作为 Pebble 中的独立 key 前缀空间保存。

这样可以：

1. 跨 CLI 调用保留待打包交易
2. 与链数据共享同一个存储引擎

### 2. mine 命令

建议增加：

1. `mine <miner-address>`

由它来触发：

1. 读取 pending tx
2. 计算总手续费
3. 创建 coinbase
4. 一起出块
5. 清空已打包交易

## 七、开发落地顺序

### 第一步：实现 Mempool 持久化

先让 pending 交易可以独立存取。

### 第二步：修改 send

让 send 不直接出块，而是把交易送进 mempool。

### 第三步：实现 mine

显式从 mempool 取交易打包出块。

### 第四步：实现手续费统计

把交易费统计出来，加到 coinbase 奖励里。

### 第五步：实现 printmempool

方便演示和调试。

## 八、CLI 如何演示

建议演示顺序：

1. `createwallet`
2. `createwallet`
3. `createblockchain <miner>`
4. `send <miner> <alice> 20 2`
5. `printmempool`
6. `mine <miner>`
7. `getbalance <miner>`
8. `getbalance <alice>`

这样老师能看到：

1. 交易先进入池中
2. 交易不是立刻成块
3. 矿工打包时拿到奖励和手续费

## 九、验收标准

1. send 进入 mempool
2. mine 打包 mempool
3. 区块第一笔是 coinbase
4. 手续费被正确计入矿工奖励
5. printmempool 可用
6. 测试通过
7. 文档同步更新

## 十、常见错误

### 错误 1：send 还直接成块

那就说明还没有真正进入交易池阶段。

### 错误 2：coinbase 奖励和手续费分离不清

这样经济模型就讲不明白。

### 错误 3：打包后 mempool 不清理

会导致重复打包。

## 十一、答辩时可以怎么讲

你可以这样解释：

1. 之前阶段是先把交易、签名、PoW链路打通
2. 这一阶段开始让出块流程更接近真实区块链运行方式
3. 交易先入池，再由矿工统一打包，并通过 coinbase + fee 获得激励

## 十二、本阶段完成后的下一步

Plan 10 完成后，下一步进入网络模拟或更复杂的链运行机制阶段。
