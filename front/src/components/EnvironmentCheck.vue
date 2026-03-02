<template>
  <div class="env-check">
    <!-- 背景网格 -->
    <div class="bg-grid"></div>

    <div class="check-container fade-in-up">
      <!-- 顶部 -->
      <div class="check-header">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
            <rect x="2" y="4" width="20" height="16" rx="2.5" stroke="var(--jm-primary-1)" stroke-width="1.5" fill="rgba(var(--jm-primary-1-rgb), 0.08)"/>
            <rect x="5" y="13" width="3" height="4" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
            <rect x="9" y="11" width="3" height="6" rx="0.5" fill="var(--jm-primary-1)"/>
            <rect x="13" y="9" width="3" height="8" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
            <rect x="17" y="7" width="3" height="10" rx="0.5" fill="var(--jm-primary-1)"/>
          </svg>
        </div>
        <div>
          <h2>环境检测</h2>
          <p class="header-desc">{{ mode === 'local' ? '检测本地编译所需的运行环境' : '检测 OpenClaw 所需的运行环境' }}</p>
        </div>
      </div>

      <!-- 检测列表 -->
      <div class="check-list">
        <!-- 扫描流光 -->
        <div v-if="isScanning" class="scan-line"></div>

        <div
          v-for="(item, i) in checkItems"
          :key="item.key"
          class="check-item"
          :class="{ scanned: item.status !== 'pending' }"
          :style="{ animationDelay: `${i * 0.12}s` }"
        >
          <!-- 状态 LED -->
          <div class="item-led" :class="item.status">
            <span class="led-dot"></span>
          </div>
          <div class="item-info">
            <span class="item-label">{{ item.label }}</span>
            <span class="item-desc">{{ item.desc }}</span>
          </div>
          <!-- 版本胶囊 / 状态 -->
          <span v-if="item.status === 'success' && item.version" class="version-pill font-mono">{{ item.version }}</span>
          <span class="item-badge" :class="item.status">
            {{ { checking: '扫描中', success: '就绪', failed: '未就绪', pending: '等待' }[item.status] }}
          </span>
        </div>
      </div>

      <!-- 结果区 -->
      <div v-if="checkComplete" class="result" :class="{ 'spring-in': checkComplete }">
        <template v-if="allReady">
          <div class="result-msg ok">
            <span class="result-led ok"><span class="led-dot"></span></span>
            <span>环境已就绪</span>
          </div>
          <button class="cta-btn" @click="$emit('passed')">
            <span class="cta-text">开始配置部署</span>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <path d="M13 5l7 7-7 7M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
        </template>
        <template v-else>
          <!-- Local 模式：一键安装 -->
          <template v-if="mode === 'local'">
            <div class="result-msg fail">
              <span class="result-led warn"><span class="led-dot"></span></span>
              <span>Node.js 环境未安装</span>
            </div>
            <button class="cta-btn" @click="installNode" :disabled="installing">
              <n-spin v-if="installing" :size="14" />
              <span class="cta-text">{{ installing ? '安装中...' : '🔧 一键安装 Node.js 环境' }}</span>
            </button>
            <button v-if="!installing" class="retry-btn" @click="retryCheck">重新检测</button>
          </template>
          <!-- Docker 模式 -->
          <template v-else>
            <div class="result-msg fail">
              <span class="result-led warn"><span class="led-dot"></span></span>
              <span>请先安装 Docker 和 Docker Compose</span>
            </div>
            <button class="retry-btn" @click="retryCheck">重新检测</button>
          </template>
        </template>
      </div>

      <!-- 返回按钮 -->
      <div class="back-row">
        <button class="back-capsule" @click="$emit('back')">
          <svg class="back-arrow" viewBox="0 0 24 24" width="14" height="14" fill="none">
            <path d="M11 19l-7-7 7-7M4 12h16" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          返回选择部署方式
        </button>
      </div>

      <!-- 安装日志 -->
      <div v-if="installing || installLogs.length" class="install-log-box">
        <div class="install-log-title">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          安装日志
        </div>
        <div class="install-log-content" ref="logBox">
          <div v-for="(log, i) in installLogs" :key="i" class="log-line">{{ log }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, defineEmits, defineProps, nextTick, watch } from 'vue'
import { NSpin } from 'naive-ui'
import { checkEnvironment, installNodeEnv, getDeployLogs } from '@/api/deploy'
import gm from '@/utils/gmssh'

const props = defineProps({
  mode: { type: String, default: 'docker' }
})
defineEmits(['passed', 'back'])

const checkComplete = ref(false)
const allReady = ref(false)
const installing = ref(false)
const installLogs = ref([])
const logBox = ref(null)
const isScanning = ref(false)

const dockerItems = [
  { key: 'docker', label: 'Docker 引擎', desc: '检测 docker 命令是否可用', status: 'pending', version: '' },
  { key: 'dockerCompose', label: 'Docker Compose', desc: '检测 docker compose 是否可用', status: 'pending', version: '' },
]
const localItems = [
  { key: 'node', label: 'Node.js', desc: '检测 node 命令是否可用', status: 'pending', version: '' },
  { key: 'pnpm', label: 'pnpm', desc: '检测 pnpm 包管理器是否可用', status: 'pending', version: '' },
]

const checkItems = reactive(props.mode === 'local' ? localItems : dockerItems)

async function runCheck() {
  checkComplete.value = false
  isScanning.value = true
  for (const item of checkItems) { item.status = 'pending'; item.version = '' }

  try {
    await sleep(400)
    const r = await checkEnvironment()

    if (props.mode === 'local') {
      checkItems[0].status = r.nodeReady ? 'success' : 'failed'
      if (r.nodeReady && r.nodeVersion) {
        checkItems[0].version = r.nodeVersion
        checkItems[0].desc = `已安装 ${r.nodeVersion}`
      }
      await sleep(350)
      checkItems[1].status = r.pnpmReady ? 'success' : 'failed'
      allReady.value = r.nodeReady && r.pnpmReady
    } else {
      checkItems[0].status = r.dockerReady ? 'success' : 'failed'
      await sleep(350)
      checkItems[1].status = r.dockerComposeReady ? 'success' : 'failed'
      allReady.value = r.allReady
    }
  } catch {
    checkItems.forEach(i => { i.status = 'failed' })
    allReady.value = false
  }
  await sleep(300)
  isScanning.value = false
  checkComplete.value = true
}

const sleep = ms => new Promise(r => setTimeout(r, ms))

function retryCheck() {
  checkItems.forEach(i => { i.status = 'pending'; i.version = '' })
  checkComplete.value = false
  installLogs.value = []
  runCheck()
}

async function installNode() {
  installing.value = true
  installLogs.value = []

  try {
    await installNodeEnv()

    const poll = setInterval(async () => {
      try {
        const logs = await getDeployLogs()
        installLogs.value = logs.logs || []
        await nextTick()
        if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight

        if (logs.finished) {
          clearInterval(poll)
          installing.value = false
          if (logs.success) {
            gm.success('Node.js 环境安装完成！')
            retryCheck()
          } else {
            gm.error('安装失败，请查看日志')
          }
        }
      } catch {}
    }, 1500)
  } catch (e) {
    installing.value = false
    gm.error('启动安装失败: ' + e.message)
  }
}

onMounted(runCheck)
</script>

<style scoped>
/* ===== 页面容器 ===== */
.env-check {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 100%; padding: 24px;
  overflow: hidden;
}

/* 背景网格 */
.bg-grid {
  position: absolute; inset: 0; z-index: 0;
  background-image:
    linear-gradient(rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
}

/* ===== 主容器（毛玻璃） ===== */
.check-container {
  position: relative; z-index: 1;
  width: 100%; max-width: 460px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--jm-glass-border);
  border-radius: 20px;
  padding: 28px 28px 20px;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 8px 40px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(255, 255, 255, 0.04);
}

/* ===== 顶部 ===== */
.check-header { display: flex; align-items: center; gap: 14px; margin-bottom: 22px; }
.header-icon {
  width: 44px; height: 44px; border-radius: 12px;
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.12), rgba(var(--jm-primary-1-rgb), 0.04));
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.15);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 0 16px rgba(var(--jm-primary-1-rgb), 0.08);
  flex-shrink: 0;
}
.check-header h2 { font-size: 16px; font-weight: 600; color: var(--jm-accent-7); margin: 0 0 3px; }
.header-desc { font-size: 12px; color: var(--jm-accent-4); margin: 0; }

/* ===== 检测列表 ===== */
.check-list {
  position: relative;
  display: flex; flex-direction: column; gap: 6px;
  margin-bottom: 22px;
}

/* 扫描流光 */
.scan-line {
  position: absolute; top: 0; left: 0; right: 0; height: 2px; z-index: 10;
  background: linear-gradient(90deg, transparent, var(--jm-primary-1), transparent);
  animation: scanDown 1.8s ease-in-out infinite;
  border-radius: 2px;
  box-shadow: 0 0 12px var(--jm-primary-1), 0 0 4px var(--jm-primary-1);
}
@keyframes scanDown {
  0% { top: 0; opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 1; }
  100% { top: 100%; opacity: 0; }
}

/* 检测卡片 */
.check-item {
  display: flex; align-items: center; gap: 12px;
  padding: 14px 16px; border-radius: 12px;
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.08);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  animation: fadeInUp 0.35s ease-out both;
  transition: all 0.3s ease;
}
.check-item.scanned {
  background: rgba(var(--jm-accent-1-rgb), 0.45);
  border-color: rgba(var(--jm-accent-2-rgb), 0.12);
  transform: translateX(2px);
}
.check-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

/* LED 状态灯 */
.item-led {
  width: 22px; height: 22px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.item-led .led-dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--jm-accent-3);
  transition: all 0.4s ease;
}
.item-led.checking .led-dot {
  background: var(--jm-primary-1);
  box-shadow: 0 0 0 3px rgba(var(--jm-primary-1-rgb), 0.15), 0 0 8px rgba(var(--jm-primary-1-rgb), 0.3);
  animation: pulse 1.2s ease-in-out infinite;
}
.item-led.success .led-dot {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.12), 0 0 8px rgba(34, 197, 94, 0.25);
}
.item-led.failed .led-dot {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.12), 0 0 8px rgba(239, 68, 68, 0.25);
}
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.3); opacity: 0.7; }
}

.item-info { flex: 1; display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.item-label { font-size: 13px; font-weight: 500; color: var(--jm-accent-7); }
.item-desc { font-size: 11px; color: var(--jm-accent-4); }

/* 版本胶囊 */
.version-pill {
  font-size: 10px; letter-spacing: 0.03em;
  padding: 2px 8px; border-radius: 6px;
  background: rgba(34, 197, 94, 0.08);
  border: 1px solid rgba(34, 197, 94, 0.12);
  color: #22c55e; font-weight: 500;
  flex-shrink: 0;
}

/* 状态标签 */
.item-badge {
  font-size: 10px; font-weight: 600; flex-shrink: 0;
  padding: 2px 8px; border-radius: 6px;
  letter-spacing: 0.02em;
}
.item-badge.pending { color: var(--jm-accent-4); background: rgba(var(--jm-accent-1-rgb), 0.3); }
.item-badge.checking { color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.08); }
.item-badge.success { color: #22c55e; background: rgba(34, 197, 94, 0.08); }
.item-badge.failed { color: #ef4444; background: rgba(239, 68, 68, 0.08); }

/* ===== 结果区 ===== */
.result { display: flex; flex-direction: column; align-items: center; gap: 14px; }
.result.spring-in { animation: springUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both; }
@keyframes springUp {
  from { opacity: 0; transform: translateY(20px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.result-msg { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 500; color: var(--jm-accent-6); }
.result-led { width: 16px; height: 16px; display: flex; align-items: center; justify-content: center; }
.result-led .led-dot { width: 8px; height: 8px; border-radius: 50%; }
.result-led.ok .led-dot {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.12), 0 0 10px rgba(34, 197, 94, 0.3);
  animation: breathe 2s ease-in-out infinite;
}
.result-led.warn .led-dot {
  background: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.12), 0 0 10px rgba(245, 158, 11, 0.3);
  animation: breathe 2s ease-in-out infinite;
}
@keyframes breathe {
  0%, 100% { box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.12), 0 0 8px rgba(34, 197, 94, 0.2); }
  50% { box-shadow: 0 0 0 5px rgba(34, 197, 94, 0.08), 0 0 16px rgba(34, 197, 94, 0.35); }
}

/* CTA 按钮 */
.cta-btn {
  width: 100%; height: 42px;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  font-size: 14px; font-weight: 500; color: #fff;
  background: linear-gradient(135deg, var(--jm-primary-1), var(--jm-primary-2));
  border: none; border-radius: 12px; cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255,255,255,0.15),
    0 4px 16px rgba(var(--jm-primary-1-rgb), 0.25),
    0 1px 3px rgba(0,0,0,0.1);
  transition: all 0.2s;
}
.cta-btn:hover {
  transform: translateY(-1px);
  box-shadow:
    inset 0 1px 0 rgba(255,255,255,0.15),
    0 6px 24px rgba(var(--jm-primary-1-rgb), 0.35),
    0 2px 6px rgba(0,0,0,0.12);
}
.cta-btn:active {
  transform: translateY(1px);
  box-shadow:
    inset 0 2px 4px rgba(0,0,0,0.15),
    0 1px 4px rgba(var(--jm-primary-1-rgb), 0.15);
}
.cta-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }

/* 重试按钮 */
.retry-btn {
  font-size: 12px; color: var(--jm-primary-2);
  background: transparent; border: none; cursor: pointer;
  padding: 4px 12px; border-radius: 6px;
  transition: all 0.2s;
}
.retry-btn:hover { background: rgba(var(--jm-primary-1-rgb), 0.06); }

/* ===== 返回按钮（胶囊） ===== */
.back-row { display: flex; justify-content: center; margin-top: 16px; }
.back-capsule {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px 6px 10px;
  border-radius: 20px;
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.2);
  background: transparent;
  color: var(--jm-accent-4); font-size: 12px;
  cursor: pointer; transition: all 0.2s;
}
.back-capsule:hover {
  color: var(--jm-primary-2);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  background: rgba(var(--jm-primary-1-rgb), 0.03);
}
.back-arrow { transition: transform 0.2s; }
.back-capsule:hover .back-arrow { animation: arrowBounce 0.4s ease; }
@keyframes arrowBounce {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(-4px); }
}

/* ===== 安装日志 ===== */
.install-log-box {
  margin-top: 16px;
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.15);
  border-radius: 10px; overflow: hidden;
  background: rgba(var(--jm-accent-1-rgb), 0.2);
}
.install-log-title {
  padding: 6px 12px;
  font-size: 10px; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase;
  color: var(--jm-accent-4);
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border-bottom: 1px solid rgba(var(--jm-accent-2-rgb), 0.1);
  display: flex; align-items: center; gap: 6px;
}
.install-log-content {
  max-height: 200px; overflow-y: auto;
  padding: 8px 12px;
  font-family: var(--jm-font-mono); font-size: 10px; line-height: 1.6;
  color: var(--jm-accent-5);
}
.install-log-content::-webkit-scrollbar { width: 4px; }
.install-log-content::-webkit-scrollbar-thumb { background: var(--jm-accent-3); border-radius: 2px; }
.log-line { white-space: pre-wrap; word-break: break-all; }

@keyframes fadeInUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
</style>
