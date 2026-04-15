import type {
  MultiSigOutputView,
  NodeStatus,
  WalletView,
} from '../types'

export type CommandPreset = {
  id: string
  label: string
  description: string
  command: string
  ready: boolean
}

export type CommandPresetGroup = {
  id: string
  title: string
  presets: CommandPreset[]
}

type BuildConsolePresetInput = {
  wallets: WalletView[]
  nodes: NodeStatus[]
  multiSigOutputs: MultiSigOutputView[]
}

export function buildConsolePresetGroups({
  wallets,
  nodes,
  multiSigOutputs,
}: BuildConsolePresetInput): CommandPresetGroup[] {
  const firstWallet = wallets[0]?.address ?? '<wallet-address>'
  const secondWallet = wallets[1]?.address ?? '<target-wallet>'
  const thirdWallet = wallets[2]?.address ?? '<wallet-3>'
  const firstNode = nodes[0]?.address ?? '<node-address>'
  const secondNode = nodes[1]?.address ?? '<seed-address>'
  const firstMultiSig = multiSigOutputs[0]
  const multiSigSigners =
    firstMultiSig?.participants.join(',') ?? `${firstWallet},${secondWallet}`

  return [
    {
      id: 'base',
      title: '基础',
      presets: [
        preset(
          'createwallet',
          '创建钱包',
          '生成一个新钱包地址。',
          'createwallet',
          true,
        ),
        preset(
          'listaddresses',
          '列出钱包',
          '查看当前本地钱包地址。',
          'listaddresses',
          true,
        ),
        preset(
          'createblockchain',
          '初始化主链',
          '用首个钱包创建区块链。',
          `createblockchain ${firstWallet}`,
          wallets.length > 0,
        ),
        preset(
          'printchain',
          '打印区块链',
          '查看从 tip 到 genesis 的链结构。',
          'printchain',
          true,
        ),
        preset(
          'printmempool',
          '查看 Mempool',
          '输出当前待打包交易。',
          'printmempool',
          true,
        ),
        preset(
          'showevents',
          '最近链事件',
          '查看最近 10 条链事件。',
          'showevents 10',
          true,
        ),
        preset(
          'showreorg',
          '最近重组',
          '查看最近一次重组摘要。',
          'showreorg',
          true,
        ),
      ],
    },
    {
      id: 'wallets',
      title: '钱包与脚本',
      presets: [
        preset(
          'getbalance',
          '查询余额',
          '查看首个钱包余额。',
          `getbalance ${firstWallet}`,
          wallets.length > 0,
        ),
        preset(
          'showscript-p2pkh',
          '查看 P2PKH 脚本',
          '输出首个钱包的标准 P2PKH 锁定脚本。',
          `showscript ${firstWallet}`,
          wallets.length > 0,
        ),
        preset(
          'showscript-p2pk',
          '查看 P2PK 脚本',
          '输出首个钱包的 P2PK 锁定脚本。',
          `showscript ${firstWallet} p2pk`,
          wallets.length > 0,
        ),
        preset(
          'showscript-multisig',
          '查看多签脚本',
          '输出 2-of-2 教学型多签脚本。',
          `showscript ${firstWallet},${secondWallet} multisig 2`,
          wallets.length > 1,
        ),
      ],
    },
    {
      id: 'transactions',
      title: '交易',
      presets: [
        preset(
          'send',
          '普通交易',
          '从首个钱包向第二个钱包转账。',
          `send ${firstWallet} ${secondWallet} 10 1`,
          wallets.length > 1,
        ),
        preset(
          'sendp2pk',
          'P2PK 交易',
          '构造一个主输出为 P2PK 的交易。',
          `sendp2pk ${firstWallet} ${secondWallet} 10 1`,
          wallets.length > 1,
        ),
        preset(
          'sendmultisig',
          '创建多签输出',
          '构造一个 2-of-2 多签输出。',
          `sendmultisig ${firstWallet} 2 ${firstWallet},${secondWallet} 12 1`,
          wallets.length > 1,
        ),
        preset(
          'spendmultisig',
          '花费多签输出',
          '花费首个未花费多签输出。',
          `spendmultisig ${multiSigSigners} ${firstMultiSig?.txid ?? '<source-txid>'} ${firstMultiSig?.out ?? 0} ${thirdWallet} 8 1`,
          Boolean(firstMultiSig) && wallets.length > 2,
        ),
        preset(
          'mine',
          '打包挖矿',
          '将当前 Mempool 打进新区块。',
          `mine ${firstWallet}`,
          wallets.length > 0,
        ),
      ],
    },
    {
      id: 'nodes',
      title: '节点',
      presets: [
        preset(
          'startnode',
          '启动节点',
          '启动一个自动分配端口的 GUI 节点。',
          'startnode 127.0.0.1:0',
          true,
        ),
        preset(
          'nodes',
          '查看节点列表',
          '列出当前 GUI 托管节点。',
          'nodes',
          true,
        ),
        preset(
          'connectnode',
          '连接节点',
          '让首个节点连接到第二个节点。',
          `connectnode ${firstNode} ${secondNode}`,
          nodes.length > 1,
        ),
        preset(
          'nodeinit',
          '初始化节点链',
          '给首个节点初始化本地链。',
          `nodeinit ${firstNode} ${firstWallet}`,
          nodes.length > 0 && wallets.length > 0,
        ),
        preset(
          'nodesend',
          '节点发交易',
          '通过首个节点发送一笔交易。',
          `nodesend ${firstNode} ${firstWallet} ${secondWallet} 10 1`,
          nodes.length > 0 && wallets.length > 1,
        ),
        preset(
          'nodemine',
          '节点挖矿',
          '让首个节点打包其待处理交易。',
          `nodemine ${firstNode}`,
          nodes.length > 0,
        ),
      ],
    },
    {
      id: 'demos',
      title: '流程',
      presets: [
        preset(
          'runnetdemo',
          '快速同步',
          '运行双节点同步与广播流程。',
          'runnetdemo',
          true,
        ),
        preset(
          'runreorgdemo',
          '重组流程',
          '运行双节点分叉 / 重组流程。',
          'runreorgdemo',
          true,
        ),
        preset(
          'runpartitiondemo',
          '分区合流',
          '运行三节点分区 / 合流流程。',
          'runpartitiondemo',
          true,
        ),
        preset(
          'simfork',
          '最长链接管',
          'CLI 层演示更长分叉接管主链。',
          `simfork ${firstWallet} 2`,
          wallets.length > 0,
        ),
        preset(
          'simreorg',
          'Mempool 恢复',
          'CLI 层演示重组后的交易恢复。',
          `simreorg ${firstWallet} ${secondWallet} 20 1`,
          wallets.length > 1,
        ),
      ],
    },
  ]
}

export function uniqueRecentCommands(commands: string[], limit = 8): string[] {
  const deduped = new Set<string>()
  const result: string[] = []

  for (const command of commands) {
    const normalized = command.trim()
    if (!normalized || deduped.has(normalized)) {
      continue
    }
    deduped.add(normalized)
    result.push(normalized)
    if (result.length >= limit) {
      break
    }
  }

  return result
}

function preset(
  id: string,
  label: string,
  description: string,
  command: string,
  ready: boolean,
): CommandPreset {
  return {
    id,
    label,
    description,
    command,
    ready,
  }
}
