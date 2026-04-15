import type { ChainEventView, NodeStatus } from '../types'

export type TopologyNodeView = {
  address: string
  shortAddress: string
  shortTipHash: string
  peerCount: number
  connectedPeers: string[]
  initialized: boolean
  height: number
  mempoolCount: number
  orphanCount: number
  hasReorg: boolean
  isMiner: boolean
}

export type TopologyLinkView = {
  key: string
  from: string
  to: string
  fromShort: string
  toShort: string
  mutual: boolean
}

export type NetworkTimelineItem = {
  id: string
  timestamp: string
  source: string
  kind: string
  detail: string
  tone: 'success' | 'warning' | 'info' | 'default'
  order: number
}

export function shortAddress(value: string, head = 8, tail = 6): string {
  if (!value || value.length <= head + tail + 3) {
    return value || '(none)'
  }
  return `${value.slice(0, head)}...${value.slice(-tail)}`
}

export function shortHash(value: string, head = 10, tail = 8): string {
  if (!value || value.length <= head + tail + 3) {
    return value || '(none)'
  }
  return `${value.slice(0, head)}...${value.slice(-tail)}`
}

export function buildTopology(nodes: NodeStatus[]) {
  const nodeMap = new Map(nodes.map((node) => [node.address, node]))

  const topologyNodes: TopologyNodeView[] = nodes
    .map((node) => {
      const connectedPeers = node.peers.filter((peer) => peer !== node.address && nodeMap.has(peer))
      return {
        address: node.address,
        shortAddress: shortAddress(node.address),
        shortTipHash: shortHash(node.tipHash),
        peerCount: connectedPeers.length,
        connectedPeers,
        initialized: node.initialized,
        height: node.height,
        mempoolCount: node.mempoolCount,
        orphanCount: node.orphanCount,
        hasReorg: Boolean(node.lastReorg),
        isMiner: Boolean(node.minerAddress),
      }
    })
    .sort((left, right) => left.address.localeCompare(right.address))

  const linkMap = new Map<string, TopologyLinkView>()
  for (const node of nodes) {
    for (const peer of node.peers) {
      if (peer === node.address || !nodeMap.has(peer)) {
        continue
      }
      const [from, to] = [node.address, peer].sort((left, right) => left.localeCompare(right))
      const key = `${from}::${to}`
      if (!linkMap.has(key)) {
        linkMap.set(key, {
          key,
          from,
          to,
          fromShort: shortAddress(from),
          toShort: shortAddress(to),
          mutual: false,
        })
      }
    }
  }

  const links = [...linkMap.values()]
    .map((link) => ({
      ...link,
      mutual: Boolean(nodeMap.get(link.from)?.peers.includes(link.to) && nodeMap.get(link.to)?.peers.includes(link.from)),
    }))
    .sort((left, right) => left.key.localeCompare(right.key))

  const initializedNodes = nodes.filter((node) => node.initialized && node.tipHash)
  const uniqueTips = [...new Set(initializedNodes.map((node) => node.tipHash))]
  const converged = initializedNodes.length > 1 && uniqueTips.length === 1
  const isolatedCount = topologyNodes.filter((node) => node.peerCount === 0).length

  return {
    nodes: topologyNodes,
    links,
    converged,
    isolatedCount,
    uniqueTipCount: uniqueTips.length,
    sharedTipHash: converged ? shortHash(uniqueTips[0]) : '',
  }
}

export function buildTimeline(nodes: NodeStatus[], chainEvents: ChainEventView[], limit = 18): NetworkTimelineItem[] {
  const timeline: NetworkTimelineItem[] = []

  for (const [index, event] of chainEvents.entries()) {
    timeline.push({
      id: `chain-${event.timestamp}-${index}`,
      timestamp: event.timestamp,
      source: '主链',
      kind: event.kind,
      detail: event.summary,
      tone: toneForKind(event.kind),
      order: parseEventTime(event.timestamp),
    })
  }

  for (const node of nodes) {
    for (const [index, event] of (node.recentEvents ?? []).entries()) {
      timeline.push({
        id: `${node.address}-${event.timestamp}-${index}`,
        timestamp: event.timestamp,
        source: shortAddress(node.address),
        kind: event.kind,
        detail: event.detail,
        tone: toneForKind(event.kind),
        order: parseEventTime(event.timestamp),
      })
    }
  }

  return timeline.sort((left, right) => right.order - left.order).slice(0, limit)
}

function toneForKind(kind: string): NetworkTimelineItem['tone'] {
  switch (kind) {
    case 'mine':
    case 'block_import':
    case 'orphan_resolved':
    case 'tip_announce':
    case 'chain_init':
    case 'genesis':
      return 'success'
    case 'reorg':
    case 'orphan':
    case 'parent_request':
      return 'warning'
    case 'connect':
    case 'listen':
    case 'version':
    case 'addr':
    case 'tx_receive':
    case 'tx_submit':
    case 'block_receive':
    case 'peer':
      return 'info'
    default:
      return 'default'
  }
}

function parseEventTime(timestamp: string): number {
  const value = Date.parse(timestamp)
  return Number.isNaN(value) ? 0 : value
}
