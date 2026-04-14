# Plan 26：第二种脚本模板 P2PK

## 一、这一阶段要解决什么问题

到 Plan 25 为止，系统虽然已经有 Script / OP 虚拟机，但真正落地使用的脚本模板仍然只有一类：

1. P2PKH

这意味着系统已经具备脚本执行能力，但还不能很好地证明“脚本系统是可扩展的”。

Plan 26 的目标，就是在现有 Script VM 上继续加入第二种真实可用的脚本模板：

1. P2PK

也就是“直接付款到公钥”的脚本形式。

## 二、为什么这一阶段值得做

因为它能直接回答一个很关键的问题：

1. 你的 Script VM 是不是只能跑一种固定模板？

如果只能跑 P2PKH，那老师可能会觉得脚本系统只是“为现有交易逻辑换了一层壳”。  
而一旦加入 P2PK，就能更明确地说明：

1. 锁定脚本可以变化
2. 解锁脚本也可以变化
3. Script VM 确实在解释不同模板

## 三、本阶段范围

### 要做的内容

1. 新增 P2PK 锁定脚本构造函数
2. 新增 P2PK 解锁脚本构造函数
3. 新增 P2PK 输出构造逻辑
4. 让 Script VM 能正确验证 P2PK 输入输出组合
5. 新增 CLI 命令 `sendp2pk`
6. 扩展 `showscript`，支持展示 `p2pk` 模板
7. 补测试与文档

### 这一阶段不做的内容

1. 不实现多重签名
2. 不实现 P2SH
3. 不实现 SegWit / Taproot
4. 不实现复杂自定义脚本编辑器

## 四、核心理论讲解

### 1. 什么是 P2PK

P2PK 就是“付款到公钥”。  
和 P2PKH 相比，它不先保存公钥哈希，而是直接在锁定脚本中保存完整公钥。  
验证时只需要：

1. 解锁脚本提供签名
2. 锁定脚本提供公钥
3. `OP_CHECKSIG` 验证签名是否匹配该公钥

### 2. P2PK 和 P2PKH 的区别

P2PKH：

1. 输出里保存公钥哈希
2. 输入里提供签名 + 公钥
3. 通过 `OP_DUP OP_HASH160 OP_EQUALVERIFY OP_CHECKSIG` 完成验证

P2PK：

1. 输出里直接保存公钥
2. 输入里只需要签名
3. 通过 `<pubkey> OP_CHECKSIG` 完成验证

### 3. 为什么 P2PK 对课程项目有价值

因为它既比 P2PKH 简单，又能很好地证明：

1. 交易锁定规则不是唯一的
2. 你的脚本引擎能解释不同模板
3. 同一条链可以容纳不同脚本输出形式

## 五、代码结构讲解

本阶段主要涉及：

1. `internal/blockchain/script.go`
2. `internal/blockchain/transaction.go`
3. `internal/blockchain/script_test.go`
4. `internal/cli/app.go`
5. `internal/cli/app_test.go`
6. `docs/plan/plan26.md`

## 六、实现步骤

### 第一步：在脚本模块中增加 P2PK 模板

补充：

1. `NewP2PKLockingScript`
2. `NewP2PKUnlockingScript`
3. P2PK 脚本结构提取函数

### 第二步：在交易输出层增加 P2PK 输出支持

让交易主输出不再只能是 P2PKH，还可以是 P2PK。

### 第三步：在签名与验证中区分模板

签名时根据被引用输出脚本模板决定：

1. P2PKH 输入脚本写什么
2. P2PK 输入脚本写什么

验证时同样根据模板走不同提取逻辑，但底层仍由统一 Script VM 执行。

### 第四步：补 CLI 演示入口

增加：

1. `sendp2pk`
2. `showscript <address> p2pk`

## 七、CLI 如何演示

建议演示顺序：

1. `createwallet`
2. `createwallet`
3. `showscript <address> p2pkh`
4. `showscript <address> p2pk`
5. `createblockchain <miner-address>`
6. `sendp2pk <from> <to> 20 1`
7. `mine <miner-address>`
8. `printchain`

演示重点：

1. 同一个系统中现在存在两种脚本模板
2. P2PK 输出的脚本结构不同于 P2PKH
3. `printchain` 中可以直接看到不同模板的脚本形式

## 八、验收标准

1. 系统支持 P2PK 锁定脚本
2. 系统支持 P2PK 解锁脚本
3. `sendp2pk` 可创建有效交易
4. `showscript` 可展示 P2PK 模板
5. 测试通过
6. 文档同步更新

## 九、常见错误

### 错误 1：只是展示 P2PK 字符串，不真正参与验证

这样不算真正支持第二种模板。

### 错误 2：P2PK 仍然沿用 P2PKH 的解锁脚本结构

这会让脚本堆栈顺序错误，验证逻辑也会失真。

### 错误 3：为了支持 P2PK 推翻现有 P2PKH 逻辑

当前更合理的方式是扩展，而不是重写。

## 十、答辩时可以怎么讲

你可以这样解释：

1. 前面阶段先把 Script VM 跑通
2. Plan 26 进一步证明它能支持不止一种模板
3. 系统现在至少能演示 P2PKH 和 P2PK 两类脚本锁定方式

## 十一、本阶段完成后的下一步

Plan 26 完成后，更适合继续推进的方向包括：

1. 多重签名
2. 更接近 Bitcoin Core 的多窗口信息组织
3. 更真实的 orphan / reorg 广播策略
4. 更细粒度的脚本与网络联合演示面板
