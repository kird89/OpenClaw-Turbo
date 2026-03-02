<template>
  <div class="cyber-page">
    <div class="cyber-container fade-in-up">
      <!-- Header -->
      <div class="cyber-header">
        <div class="header-left">
          <h2 class="page-title">
            <span v-html="icons.brainChip(20, 20)"></span>
            赛博指挥中心
          </h2>
          <span class="header-hint">实时监控你的 Agent 团队协同状态</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchAgents" :disabled="loading">
            <span :class="{ spinning: loading }" v-html="icons.refresh(16, 16)"></span>
          </button>
        </div>
      </div>

      <!-- Topology Canvas -->
      <div class="topology-viewport" ref="viewportRef">
        <!-- Grid Background -->
        <div class="grid-bg"></div>

        <!-- Scan Line Effect -->
        <div class="scan-line"></div>

        <!-- SVG Connection Lines -->
        <svg class="connections-layer" :viewBox="`0 0 ${vpWidth} ${vpHeight}`">
          <defs>
            <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" style="stop-color: var(--jm-primary-1); stop-opacity: 0.1" />
              <stop offset="50%" style="stop-color: var(--jm-primary-1); stop-opacity: 0.6" />
              <stop offset="100%" style="stop-color: var(--jm-primary-1); stop-opacity: 0.1" />
            </linearGradient>
            <filter id="glow">
              <feGaussianBlur stdDeviation="3" result="blur"/>
              <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
            </filter>
          </defs>
          <!-- Connection lines from main to specialists -->
          <g v-for="conn in connections" :key="conn.id">
            <line
              :x1="conn.x1" :y1="conn.y1"
              :x2="conn.x2" :y2="conn.y2"
              stroke="url(#lineGrad)"
              stroke-width="1.5"
              filter="url(#glow)"
            />
            <!-- Flowing data particle -->
            <circle
              r="3"
              fill="var(--jm-primary-1)"
              filter="url(#glow)"
              opacity="0.8"
            >
              <animateMotion
                :dur="(2 + Math.random() * 2) + 's'"
                repeatCount="indefinite"
                :path="`M${conn.x1},${conn.y1} L${conn.x2},${conn.y2}`"
              />
            </circle>
            <circle
              r="2"
              fill="var(--jm-primary-2)"
              filter="url(#glow)"
              opacity="0.5"
            >
              <animateMotion
                :dur="(3 + Math.random() * 2) + 's'"
                repeatCount="indefinite"
                :path="`M${conn.x2},${conn.y2} L${conn.x1},${conn.y1}`"
              />
            </circle>
          </g>
        </svg>

        <!-- Agent Nodes -->
        <div
          v-for="node in agentNodes"
          :key="node.id"
          class="agent-node"
          :class="[node.role, node.status]"
          :style="{ left: node.x + 'px', top: node.y + 'px' }"
          @click="navigateToAgent(node)"
        >
          <!-- Pulse Ring -->
          <div class="pulse-ring" :class="node.status"></div>
          <div class="pulse-ring delay" :class="node.status"></div>

          <!-- Core -->
          <div class="node-core">
            <span class="node-icon" v-html="(icons[node.avatar] || icons.robot)(node.role === 'main' ? 32 : 24, node.role === 'main' ? 32 : 24)"></span>
          </div>

          <!-- Info -->
          <div class="node-label">{{ node.name }}</div>
          <div class="node-status-text">{{ statusLabel(node.status) }}</div>
        </div>

        <!-- Empty state -->
        <div v-if="!loading && agents.length === 0" class="empty-state">
          <span>暂无 Agent，去管理页面创建一个吧</span>
        </div>
      </div>

      <!-- Bottom Stats Bar -->
      <div class="stats-bar" v-if="agents.length > 0">
        <div class="stat-item">
          <span class="stat-value">{{ agents.length }}</span>
          <span class="stat-label">Agent 总数</span>
        </div>
        <div class="stat-item">
          <span class="stat-dot idle"></span>
          <span class="stat-value">{{ countByStatus('idle') }}</span>
          <span class="stat-label">待命</span>
        </div>
        <div class="stat-item">
          <span class="stat-dot thinking"></span>
          <span class="stat-value">{{ countByStatus('thinking') }}</span>
          <span class="stat-label">思考中</span>
        </div>
        <div class="stat-item">
          <span class="stat-dot acting"></span>
          <span class="stat-value">{{ countByStatus('acting') }}</span>
          <span class="stat-label">执行中</span>
        </div>
        <div class="stat-item">
          <span class="stat-dot error"></span>
          <span class="stat-value">{{ countByStatus('error') }}</span>
          <span class="stat-label">异常</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { listAgents } from '@/api/agent'
import { icons } from '@/components/icons'
import gm from '@/utils/gmssh'

const router = useRouter()
const loading = ref(true)
const agents = ref([])
const viewportRef = ref(null)
const vpWidth = ref(800)
const vpHeight = ref(500)

function statusLabel(status) {
  const map = { idle: 'IDLE', thinking: 'THINKING...', acting: 'EXECUTING', error: 'ERROR' }
  return map[status] || status.toUpperCase()
}

function countByStatus(status) {
  return agents.value.filter(a => a.status === status).length
}

// ===== Layout Calculation =====
const agentNodes = computed(() => {
  const list = agents.value
  if (list.length === 0) return []

  const w = vpWidth.value
  const h = vpHeight.value
  const cx = w / 2
  const cy = h / 2

  const mainAgent = list.find(a => a.role === 'main')
  const specialists = list.filter(a => a.role !== 'main')

  const nodes = []

  if (mainAgent) {
    nodes.push({ ...mainAgent, x: cx - 40, y: cy - 50 })
  }

  // Distribute specialists in a circle around the main agent
  const radius = Math.min(w, h) * 0.32
  specialists.forEach((agent, i) => {
    const angle = (2 * Math.PI * i) / Math.max(specialists.length, 1) - Math.PI / 2
    const x = cx + radius * Math.cos(angle) - 35
    const y = cy + radius * Math.sin(angle) - 35
    nodes.push({ ...agent, x, y })
  })

  return nodes
})

const connections = computed(() => {
  const main = agentNodes.value.find(n => n.role === 'main')
  if (!main) return []
  return agentNodes.value
    .filter(n => n.role !== 'main')
    .map(n => ({
      id: `${main.id}-${n.id}`,
      x1: main.x + 40,
      y1: main.y + 40,
      x2: n.x + 35,
      y2: n.y + 35,
    }))
})

async function fetchAgents() {
  loading.value = true
  try {
    const res = await listAgents()
    if (res?.agents) agents.value = res.agents
  } catch (e) {
    gm.error('加载失败: ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

function navigateToAgent(node) {
  router.push('/agents')
}

function updateViewport() {
  if (viewportRef.value) {
    vpWidth.value = viewportRef.value.clientWidth
    vpHeight.value = viewportRef.value.clientHeight
  }
}

let pollTimer = null
let resizeObs = null

onMounted(async () => {
  await fetchAgents()
  await nextTick()
  updateViewport()

  resizeObs = new ResizeObserver(updateViewport)
  if (viewportRef.value) resizeObs.observe(viewportRef.value)

  pollTimer = setInterval(async () => {
    try {
      const res = await listAgents()
      if (res?.agents) agents.value = res.agents
    } catch { /* silent */ }
  }, 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (resizeObs) resizeObs.disconnect()
})
</script>

<style scoped>
.cyber-page {
  width: 100%; height: 100%;
  overflow: hidden;
  padding: 20px;
  display: flex;
}
.cyber-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

/* ===== Header ===== */
.cyber-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
}
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 18px; font-weight: 700;
  color: var(--jm-accent-7); margin: 0;
  letter-spacing: -0.01em;
}
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }

/* ===== Topology Viewport — Glass ===== */
.topology-viewport {
  flex: 1;
  position: relative;
  border-radius: 16px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  overflow: hidden;
  min-height: 400px;
  box-shadow: var(--jm-glass-inner-glow), var(--jm-shadow-elevation-2);
}

/* Grid Background — more visible */
.grid-bg {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(var(--jm-primary-1-rgb), 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--jm-primary-1-rgb), 0.04) 1px, transparent 1px);
  background-size: 40px 40px;
}

/* Scan Line — brighter */
.scan-line {
  position: absolute;
  left: 0; right: 0;
  height: 2px;
  background: linear-gradient(90deg,
    transparent 0%,
    rgba(var(--jm-primary-1-rgb), 0.15) 20%,
    rgba(var(--jm-primary-1-rgb), 0.5) 50%,
    rgba(var(--jm-primary-1-rgb), 0.15) 80%,
    transparent 100%
  );
  animation: scan 4s linear infinite;
  pointer-events: none;
}
@keyframes scan {
  0% { top: -2px; }
  100% { top: 100%; }
}

/* SVG Layer */
.connections-layer {
  position: absolute; inset: 0;
  width: 100%; height: 100%;
  pointer-events: none;
}

/* ===== Agent Nodes ===== */
.agent-node {
  position: absolute;
  display: flex; flex-direction: column; align-items: center;
  gap: 6px; cursor: pointer;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  z-index: 2;
}
.agent-node:hover { transform: scale(1.1); }

/* Pulse Rings */
.pulse-ring {
  position: absolute;
  top: 50%; left: 50%;
  width: 70px; height: 70px;
  margin: -35px 0 0 -35px;
  border-radius: 50%;
  border: 1px solid var(--jm-primary-1);
  opacity: 0;
  animation: pulse-ring 3s ease-out infinite;
  pointer-events: none;
}
.pulse-ring.delay { animation-delay: 1.5s; }
.pulse-ring.idle { border-color: #4ade80; }
.pulse-ring.thinking { border-color: #60a5fa; animation-duration: 1.5s; }
.pulse-ring.acting { border-color: #fb923c; animation-duration: 1s; }
.pulse-ring.error { border-color: #f87171; animation-duration: 0.8s; }

@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 0.6; }
  100% { transform: scale(2.2); opacity: 0; }
}

/* Main Agent larger pulse */
.agent-node.main .pulse-ring {
  width: 90px; height: 90px;
  margin: -45px 0 0 -45px;
}

/* Node Core — Glassmorphism */
.node-core {
  width: 64px; height: 64px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  border: 2px solid rgba(var(--jm-primary-1-rgb), 0.4);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px);
  position: relative; z-index: 1;
  transition: all 0.3s;
  box-shadow: var(--jm-glass-inner-glow);
}
.agent-node:hover .node-core {
  border-color: var(--jm-primary-1);
  box-shadow: 0 0 24px rgba(var(--jm-primary-1-rgb), 0.35), var(--jm-glass-inner-glow);
}

.agent-node.main .node-core {
  width: 80px; height: 80px;
  border-width: 2px;
  border-color: var(--jm-primary-1);
  box-shadow: 0 0 24px rgba(var(--jm-primary-1-rgb), 0.25),
              inset 0 0 15px rgba(var(--jm-primary-1-rgb), 0.05),
              var(--jm-glass-inner-glow);
}

/* Status-based core glow */
.agent-node.idle .node-core { border-color: rgba(74, 222, 128, 0.4); }
.agent-node.thinking .node-core {
  border-color: rgba(96, 165, 250, 0.6);
  box-shadow: 0 0 20px rgba(96, 165, 250, 0.25), var(--jm-glass-inner-glow);
  animation: think-pulse 1.5s ease-in-out infinite;
}
.agent-node.acting .node-core {
  border-color: rgba(251, 146, 60, 0.6);
  box-shadow: 0 0 20px rgba(251, 146, 60, 0.25), var(--jm-glass-inner-glow);
}
.agent-node.error .node-core {
  border-color: rgba(248, 113, 113, 0.8);
  box-shadow: 0 0 20px rgba(248, 113, 113, 0.35), var(--jm-glass-inner-glow);
  animation: error-glitch 0.5s ease-in-out infinite;
}

@keyframes think-pulse {
  0%, 100% { box-shadow: 0 0 15px rgba(96, 165, 250, 0.2), var(--jm-glass-inner-glow); }
  50% { box-shadow: 0 0 35px rgba(96, 165, 250, 0.5), var(--jm-glass-inner-glow); }
}
@keyframes error-glitch {
  0%, 100% { transform: none; }
  25% { transform: translate(1px, -1px); }
  50% { transform: translate(-1px, 1px); }
  75% { transform: translate(1px, 1px); }
}

.node-icon {
  display: flex; align-items: center; justify-content: center;
  color: var(--jm-accent-5);
}
.agent-node.main .node-icon { color: var(--jm-primary-1); }
.agent-node:hover .node-icon { color: var(--jm-primary-1); }

.node-label {
  font-size: 12px; font-weight: 600;
  color: var(--jm-accent-7);
  text-shadow: 0 0 12px rgba(0,0,0,0.9);
  white-space: nowrap;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(0,0,0,0.25);
  backdrop-filter: blur(4px); -webkit-backdrop-filter: blur(4px);
}
.node-status-text {
  font-size: 9px; font-weight: 600;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: var(--jm-accent-4);
  font-family: var(--jm-font-mono);
}
.agent-node.idle .node-status-text { color: #4ade80; text-shadow: 0 0 6px rgba(74,222,128,0.4); }
.agent-node.thinking .node-status-text { color: #60a5fa; text-shadow: 0 0 6px rgba(96,165,250,0.4); }
.agent-node.acting .node-status-text { color: #fb923c; text-shadow: 0 0 6px rgba(251,146,60,0.4); }
.agent-node.error .node-status-text { color: #f87171; text-shadow: 0 0 6px rgba(248,113,113,0.4); }

/* Empty State */
.empty-state {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  color: var(--jm-accent-4); font-size: 14px;
}

/* ===== Stats Bar — Glass ===== */
.stats-bar {
  display: flex; align-items: center; justify-content: center;
  gap: 24px; padding: 10px 20px;
  border-radius: 12px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  flex-shrink: 0;
  box-shadow: var(--jm-glass-inner-glow);
}
.stat-item { display: flex; align-items: center; gap: 6px; }
.stat-value {
  font-size: 16px; font-weight: 700;
  color: var(--jm-accent-7);
  font-family: var(--jm-font-mono);
}
.stat-label { font-size: 11px; color: var(--jm-accent-4); }
.stat-dot { width: 8px; height: 8px; border-radius: 50%; }
.stat-dot.idle { background: #4ade80; box-shadow: 0 0 6px #4ade80; }
.stat-dot.thinking { background: #60a5fa; box-shadow: 0 0 6px #60a5fa; }
.stat-dot.acting { background: #fb923c; box-shadow: 0 0 6px #fb923c; }
.stat-dot.error { background: #f87171; box-shadow: 0 0 6px #f87171; }

.fade-in-up { animation: fadeInUp 0.4s ease; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: none; } }
</style>
