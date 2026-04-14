# 答辩前检查清单

## 一、环境检查

- [ ] Go 可用
- [ ] Bun 可用
- [ ] Wails 可用
- [ ] GUI 可启动
- [ ] 没有遗留占用中的 GUI 进程

## 二、数据准备

- [ ] 至少准备 2 个钱包地址
- [ ] 准备 1 个矿工地址
- [ ] 区块链数据目录干净或符合演示预期
- [ ] GUI 使用独立数据目录

## 三、CLI 演示检查

- [ ] `createwallet`
- [ ] `listaddresses`
- [ ] `createblockchain`
- [ ] `showscript`
- [ ] `send`
- [ ] `printmempool`
- [ ] `mine`
- [ ] `getbalance`
- [ ] `printchain`
- [ ] `simdouble`
- [ ] `runperf`

## 四、GUI 演示检查

- [ ] Dashboard 正常显示
- [ ] 钱包页面可创建钱包
- [ ] 钱包页面能显示锁定脚本
- [ ] 区块页面可浏览区块
- [ ] 区块页面能显示 `scriptSig` / `scriptPubKey`
- [ ] 交易页面可排队交易
- [ ] 挖矿按钮可工作
- [ ] 浅色/深色模式切换正常

## 五、网络演示检查

- [ ] 两个节点能启动
- [ ] 节点能发现彼此
- [ ] 交易能传播
- [ ] 区块能传播
- [ ] 节点链高度可同步

## 六、实验结果检查

- [ ] `docs/perf/latest.json` 存在
- [ ] `docs/perf/latest.md` 存在
- [ ] 结果数据是最新的
- [ ] speedup 数值可解释

## 七、答辩材料检查

- [ ] `docs/architecture.md`
- [ ] `docs/report.md`
- [ ] `docs/defense-script.md`
- [ ] `docs/plan/` 文档完整
- [ ] Script VM 演示说辞已准备好

## 八、最后确认

- [ ] 演示顺序清楚
- [ ] 突发错误时知道备用方案
- [ ] 关键命令已经提前跑过
