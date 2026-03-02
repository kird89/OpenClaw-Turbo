<template>
  <div class="channels-page">
    <div class="channels-container fade-in-up">
      <!-- 顶部 -->
      <div class="channels-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5">
              <path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9M7.8 16.2a5.5 5.5 0 010-8.4"/>
              <circle cx="12" cy="12" r="2"/>
              <path d="M16.2 7.8a5.5 5.5 0 010 8.4M19.1 4.9C23 8.8 23 15.2 19.1 19.1"/>
            </svg>
            消息频道
          </h2>
          <span class="header-hint">管理您的消息频道和连接</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchChannels()" :disabled="loading" title="刷新">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
          </button>
          <button class="add-btn" @click="showAddModal = true">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
            添加频道
          </button>
        </div>
      </div>

      <!-- 主体：双栏布局 -->
      <div class="channel-body">
        <!-- 左栏：频道网格 -->
        <div class="channel-main">
          <!-- 已接入频道 -->
          <div v-if="channels.length > 0" class="section-block">
            <div class="section-title-row">
              <h3 class="section-label">已接入频道</h3>
              <span class="section-hint">点击编辑配置</span>
            </div>
            <div class="connected-grid">
              <div v-for="ch in channels" :key="ch.key" class="connected-card" @click="editChannel(ch)">
                <div class="cc-top">
                  <div class="cc-icon" v-html="getChannelIcon(ch.key)"></div>
                  <div class="cc-info">
                    <span class="cc-name">{{ getChannelDisplayName(ch.key) }}</span>
                    <span class="cc-status" :class="ch.enabled ? 'on' : 'off'">
                      <span class="cc-dot"></span>{{ ch.enabled ? '已连接' : '已禁用' }}
                    </span>
                  </div>
                </div>
                <div class="cc-actions" @click.stop>
                  <button class="cc-btn" @click.stop="toggleCh(ch)">{{ ch.enabled ? '禁用' : '启用' }}</button>
                  <button class="cc-btn danger" @click.stop="deleteCh(ch)">删除</button>
                </div>
              </div>
            </div>
          </div>

          <!-- 可用频道 -->
          <div class="section-block">
            <div class="section-title-row">
              <h3 class="section-label">可用频道</h3>
              <span class="section-hint">连接一个新的频道</span>
            </div>
            <div class="available-grid">
              <button
                v-for="t in channelTypes"
                :key="t.key"
                class="available-card"
                :class="{ soon: !t.available, active: channels.some(c => c.key === t.key) }"
                @click="t.available && openAddForm(t.key)"
              >
                <span class="avail-icon" v-html="t.icon"></span>
                <span class="avail-name">{{ t.name }}</span>
                <span class="avail-auth">{{ t.auth }}</span>
                <span v-if="!t.available" class="avail-badge">即将</span>
                <span v-else-if="channels.some(c => c.key === t.key)" class="avail-badge connected-badge">已接入</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 右栏：监控面板 -->
        <div class="channel-sidebar">
          <div class="sidebar-stat">
            <div class="sidebar-stat-icon total">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9M7.8 16.2a5.5 5.5 0 010-8.4"/><circle cx="12" cy="12" r="2"/><path d="M16.2 7.8a5.5 5.5 0 010 8.4M19.1 4.9C23 8.8 23 15.2 19.1 19.1"/></svg>
            </div>
            <span class="sidebar-stat-val">{{ channels.length }}</span>
            <span class="sidebar-stat-label">频道总数</span>
          </div>
          <div class="sidebar-stat">
            <div class="sidebar-stat-icon connected">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            </div>
            <span class="sidebar-stat-val">{{ enabledCount }}</span>
            <span class="sidebar-stat-label">已连接</span>
          </div>
          <div class="sidebar-stat">
            <div class="sidebar-stat-icon disconnected">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
            </div>
            <span class="sidebar-stat-val">{{ disabledCount }}</span>
            <span class="sidebar-stat-label">未连接</span>
          </div>
        </div>
      </div>

      <!-- 加载 -->
      <div v-if="loading && channels.length === 0" class="loading-state">
        <div class="loading-spinner"></div>
      </div>
    </div>

    <!-- ===== 添加频道弹框 ===== -->
    <n-modal v-model:show="showAddModal" preset="card" title="添加频道" :bordered="false" size="small" style="max-width: 460px;" :mask-closable="true">
      <p class="modal-subtitle">选择要配置的频道类型</p>
      <div class="modal-type-grid">
        <button
          v-for="t in channelTypes.filter(c => c.available)"
          :key="t.key"
          class="modal-type-card"
          @click="showAddModal = false; openAddForm(t.key)"
        >
          <span class="modal-type-icon" v-html="t.icon"></span>
          <span class="modal-type-name">{{ t.name }}</span>
          <span class="modal-type-auth">{{ t.auth }}</span>
        </button>
      </div>
    </n-modal>

    <!-- ===== 企业微信配置弹框 ===== -->
    <n-modal v-model:show="showWecomModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' 企业微信'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="wecomForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="Corp ID (企业 ID)" required><n-input v-model:value="wecomForm.corpId" placeholder="输入企业 ID（如 wwxxxxxxxxxx）" /></n-form-item>
        <n-form-item label="Corp Secret (应用密钥)" required><n-input v-model:value="wecomForm.corpSecret" type="password" show-password-on="click" placeholder="应用的 Secret" /></n-form-item>
        <n-form-item label="Agent ID (应用 ID)" required><n-input-number v-model:value="wecomForm.agentId" :min="1" style="width: 100%" placeholder="1000002" /></n-form-item>
        <n-form-item label="Token (接收消息Token)" required><n-input v-model:value="wecomForm.token" placeholder="应用回调 Token" /></n-form-item>
        <n-form-item label="Encoding AES Key" required><n-input v-model:value="wecomForm.encodingAESKey" placeholder="消息加密密钥" /></n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveWecom" :loading="saving" :disabled="!wecomForm.corpId || !wecomForm.corpSecret || !wecomForm.agentId || !wecomForm.token || !wecomForm.encodingAESKey">保存</n-button>
          </div>
        </div>
      </template>
    </n-modal>

    <!-- ===== QQ 机器人配置弹框 ===== -->
    <n-modal v-model:show="showQQModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' QQ 机器人'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="qqForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="App ID (机器人 AppID)" required><n-input v-model:value="qqForm.appId" placeholder="你的 AppID" /></n-form-item>
        <n-form-item label="Client Secret (AppSecret)" required><n-input v-model:value="qqForm.clientSecret" type="password" show-password-on="click" placeholder="你的 AppSecret" /></n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveQQBot" :loading="saving" :disabled="!qqForm.appId || !qqForm.clientSecret">保存</n-button>
          </div>
        </div>
      </template>
    </n-modal>

    <!-- ===== 钉钉配置弹框 ===== -->
    <n-modal v-model:show="showDingtalkModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' 钉钉'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="dingtalkForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="AgentID" required><n-input v-model:value="dingtalkForm.agentId" placeholder="123456789" /></n-form-item>
        <n-form-item label="Client ID (AppKey)" required><n-input v-model:value="dingtalkForm.clientId" placeholder="dingxxxxxx" /></n-form-item>
        <n-form-item label="Client Secret" required><n-input v-model:value="dingtalkForm.clientSecret" type="password" show-password-on="click" placeholder="应用 AppSecret" /></n-form-item>
        <n-form-item label="Robot Code" required><n-input v-model:value="dingtalkForm.robotCode" placeholder="dingxxxxxx" /></n-form-item>
        <n-form-item label="Corp ID (企业 ID)" required><n-input v-model:value="dingtalkForm.corpId" placeholder="dingxxxxxx" /></n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveDingTalk" :loading="saving" :disabled="!dingtalkForm.clientId || !dingtalkForm.clientSecret || !dingtalkForm.robotCode || !dingtalkForm.corpId || !dingtalkForm.agentId">保存</n-button>
          </div>
        </div>
      </template>
    </n-modal>

    <!-- ===== 飞书配置弹框 ===== -->
    <n-modal v-model:show="showFeishuModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' 飞书'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="feishuForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="App ID" required><n-input v-model:value="feishuForm.appId" placeholder="cli_xxx" /></n-form-item>
        <n-form-item label="App Secret" required><n-input v-model:value="feishuForm.appSecret" type="password" show-password-on="click" placeholder="应用 Secret" /></n-form-item>
        <n-form-item label="Bot Name" required><n-input v-model:value="feishuForm.botName" placeholder="我的AI助手" /></n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveFeishu" :loading="saving" :disabled="!feishuForm.appId || !feishuForm.appSecret || !feishuForm.botName">保存</n-button>
          </div>
        </div>
      </template>
    </n-modal>

    <!-- ===== Discord 配置弹框 ===== -->
    <n-modal v-model:show="showDiscordModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' Discord'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="discordForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="机器人 Token" required>
          <n-input v-model:value="discordForm.token" type="password" show-password-on="click" placeholder="MTQ3Njg2..." />
        </n-form-item>
        <n-form-item label="服务器 ID (Guild ID)" required>
          <n-input v-model:value="discordForm.guildId" placeholder="1476867320373837877" />
          <template #feedback>
            <span class="env-hint">右键服务器名称 → 复制服务器 ID（需开启开发者模式）</span>
          </template>
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">返回</n-button>
            <n-button type="primary" @click="saveDiscord" :loading="saving" :disabled="!discordForm.token || !discordForm.guildId.trim()">保存并连接</n-button>
          </div>
        </div>
      </template>
    </n-modal>

    <!-- ===== Telegram 配置弹框 ===== -->
    <n-modal v-model:show="showTelegramModal" preset="card" :title="(editingChannel ? '编辑' : '新增') + ' Telegram'" :bordered="false" size="small" style="max-width: 520px;">
      <n-form :model="telegramForm" label-placement="top" class="channel-form" size="medium">
        <n-form-item label="机器人令牌" required>
          <n-input v-model:value="telegramForm.botToken" type="password" show-password-on="click" placeholder="123456:ABC-DEF..." />
          <template #feedback>
            <span class="env-hint">环境变量: TELEGRAM_BOT_TOKEN</span>
          </template>
        </n-form-item>
        <n-form-item label="允许的用户 ID" required>
          <n-input v-model:value="telegramForm.allowFromText" placeholder="例如 123456789, 987654321" />
          <template #feedback>
            <span class="env-hint">允许使用机器人的用户 ID 列表（逗号分隔）。出于安全考虑，此项为必填。</span>
          </template>
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置？</a>
          <div class="modal-actions">
            <n-button quaternary size="small" @click="cancelForm">返回</n-button>
            <n-button type="primary" @click="saveTelegram" :loading="saving" :disabled="!telegramForm.botToken || !telegramForm.allowFromText.trim()">保存并连接</n-button>
          </div>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { NForm, NFormItem, NInput, NInputNumber, NButton, NTooltip, NModal } from 'naive-ui'
import { getChannels, saveChannel, deleteChannel, toggleChannel } from '@/api/channel'
import gm from '@/utils/gmssh'
import cache from '@/stores/cache'

const loading = ref(true)
const saving = ref(false)
const channels = ref([])
const editingChannel = ref(null)


// 弹框状态
const showAddModal = ref(false)
const showWecomModal = ref(false)
const showQQModal = ref(false)
const showDingtalkModal = ref(false)
const showFeishuModal = ref(false)
const showDiscordModal = ref(false)
const showTelegramModal = ref(false)

// 统计
const enabledCount = computed(() => channels.value.filter(c => c.enabled).length)
const disabledCount = computed(() => channels.value.filter(c => !c.enabled).length)

const wecomForm = ref({ corpId: '', corpSecret: '', agentId: null, token: '', encodingAESKey: '' })
const qqForm = ref({ appId: '', clientSecret: '' })
const dingtalkForm = ref({ clientId: '', clientSecret: '', robotCode: '', corpId: '', agentId: '' })
const discordForm = ref({ token: '', guildId: '' })
const telegramForm = ref({ botToken: '', allowFromText: '' })
const feishuForm = ref({ appId: '', appSecret: '', botName: '' })

const channelTypes = [
  { key: 'wecom-app', name: '企业微信', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><rect x="3" y="3" width="18" height="18" rx="3" stroke="currentColor" stroke-width="1.3"/><path d="M8 10.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM14 10.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM7 14.5s1.5 2 5 2 5-2 5-2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>' },
  { key: 'qqbot', name: 'QQ 机器人', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.3"/><circle cx="9" cy="10" r="1.5" fill="currentColor"/><circle cx="15" cy="10" r="1.5" fill="currentColor"/><path d="M8 15s2 2 4 2 4-2 4-2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>' },
  { key: 'dingtalk', name: '钉钉', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M12 2L2 7l10 5 10-5-10-5z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/><path d="M2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>' },
  { key: 'feishu', name: '飞书 / Lark', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M4 4l16 8-16 8V4z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>' },
  { key: 'telegram', name: 'Telegram', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>' },
  { key: 'discord', name: 'Discord', available: true, auth: '令牌',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M9.5 11.5a1 1 0 100-2 1 1 0 000 2zM14.5 11.5a1 1 0 100-2 1 1 0 000 2z" fill="currentColor"/><path d="M5.5 16c1 2 3.5 3 6.5 3s5.5-1 6.5-3M8.5 6A13.7 13.7 0 005 8.5S2.5 13 3 18c1.5 1.5 4 2.5 4 2.5l1-2a11 11 0 01-2-1M15.5 6A13.7 13.7 0 0119 8.5S21.5 13 21 18c-1.5 1.5-4 2.5-4 2.5l-1-2a11 11 0 002-1" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
  { key: 'whatsapp', name: 'WhatsApp', available: false, auth: '二维码',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z" stroke="currentColor" stroke-width="1.3"/></svg>' },
  { key: 'imessage', name: 'iMessage', available: false, auth: '插件',
    icon: '<svg viewBox="0 0 24 24" width="28" height="28" fill="none"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke="currentColor" stroke-width="1.3"/></svg>' },
]

function getChannelDisplayName(key) {
  const t = channelTypes.find(c => c.key === key)
  return t ? t.name : key
}

function getChannelIcon(key) {
  const t = channelTypes.find(c => c.key === key)
  return t ? t.icon : ''
}

// 卡片字段
const revealedFields = ref({})

function getChannelFields(ch) {
  if (ch.key === 'wecom-app') {
    return [
      { label: 'Corp ID', value: String(ch.corpId || ''), masked: false },
      { label: 'Agent ID', value: String(ch.agentId || ''), masked: false },
      { label: 'Corp Secret', value: String(ch.corpSecret || ''), masked: true },
      { label: 'Token', value: String(ch.token || ''), masked: true },
      { label: 'AES Key', value: String(ch.encodingAESKey || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'qqbot') {
    return [
      { label: 'App ID', value: String(ch.appId || ''), masked: false },
      { label: 'Client Secret', value: String(ch.clientSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'dingtalk') {
    return [
      { label: 'Client ID', value: String(ch.clientId || ''), masked: false },
      { label: 'Robot Code', value: String(ch.robotCode || ''), masked: false },
      { label: 'Corp ID', value: String(ch.corpId || ''), masked: false },
      { label: 'Agent ID', value: String(ch.agentId || ''), masked: false },
      { label: 'Client Secret', value: String(ch.clientSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'feishu') {
    const acc = ch.accounts?.main || {}
    return [
      { label: 'App ID', value: String(acc.appId || ''), masked: false },
      { label: 'Bot Name', value: String(acc.botName || ''), masked: false },
      { label: 'App Secret', value: String(acc.appSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'telegram') {
    const fields = [
      { label: 'Bot Token', value: String(ch.botToken || ''), masked: true },
    ]
    if (ch.allowFrom && Array.isArray(ch.allowFrom)) {
      const ids = ch.allowFrom.filter(x => x !== '*')
      if (ids.length) fields.push({ label: '允许用户', value: ids.join(', '), masked: false })
    }
    return fields.filter(f => f.value)
  }
  if (ch.key === 'discord') {
    const guildIds = ch.guilds ? Object.keys(ch.guilds) : []
    return [
      { label: 'Token', value: String(ch.token || ''), masked: true },
      { label: '服务器 ID', value: guildIds.join(', ') || '-', masked: false },
    ].filter(f => f.value)
  }
  return []
}

function maskValue(val) {
  if (!val || val.length <= 4) return '****'
  return val.slice(0, 3) + '****' + val.slice(-3)
}

function displayValue(field, channelKey) {
  if (field.masked && !revealedFields.value[channelKey + field.label]) return maskValue(field.value)
  return field.value
}

function toggleReveal(key) { revealedFields.value[key] = !revealedFields.value[key] }

function openAddForm(key) {
  editingChannel.value = null
  resetForms()
  const modalMap = {
    'wecom-app': showWecomModal,
    'qqbot': showQQModal,
    'dingtalk': showDingtalkModal,
    'feishu': showFeishuModal,
    'discord': showDiscordModal,
    'telegram': showTelegramModal,
  }
  if (modalMap[key]) modalMap[key].value = true
}

function resetForms() {
  wecomForm.value = { corpId: '', corpSecret: '', agentId: null, token: '', encodingAESKey: '' }
  qqForm.value = { appId: '', clientSecret: '' }
  dingtalkForm.value = { clientId: '', clientSecret: '', robotCode: '', corpId: '', agentId: '' }
  feishuForm.value = { appId: '', appSecret: '', botName: '' }
  discordForm.value = { token: '', guildId: '' }
  telegramForm.value = { botToken: '', allowFromText: '' }
}

function cancelForm() {
  showWecomModal.value = false
  showQQModal.value = false
  showDingtalkModal.value = false
  showFeishuModal.value = false
  showDiscordModal.value = false
  showTelegramModal.value = false
  editingChannel.value = null
  resetForms()
}

function editChannel(ch) {
  editingChannel.value = ch
  if (ch.key === 'wecom-app') {
    wecomForm.value = { corpId: ch.corpId || '', corpSecret: ch.corpSecret || '', agentId: ch.agentId || null, token: ch.token || '', encodingAESKey: ch.encodingAESKey || '' }
    showWecomModal.value = true
  } else if (ch.key === 'dingtalk') {
    dingtalkForm.value = { clientId: ch.clientId || '', clientSecret: ch.clientSecret || '', robotCode: ch.robotCode || '', corpId: ch.corpId || '', agentId: String(ch.agentId || '') }
    showDingtalkModal.value = true
  } else if (ch.key === 'qqbot') {
    qqForm.value = { appId: ch.appId || '', clientSecret: ch.clientSecret || '' }
    showQQModal.value = true
  } else if (ch.key === 'feishu') {
    const acc = ch.accounts?.main || {}
    feishuForm.value = { appId: acc.appId || '', appSecret: acc.appSecret || '', botName: acc.botName || '' }
    showFeishuModal.value = true
  } else if (ch.key === 'discord') {
    const guildIds = ch.guilds ? Object.keys(ch.guilds) : []
    discordForm.value = { token: ch.token || '', guildId: guildIds[0] || '' }
    showDiscordModal.value = true
  } else if (ch.key === 'telegram') {
    const ids = Array.isArray(ch.allowFrom) ? ch.allowFrom.filter(x => x !== '*') : []
    telegramForm.value = { botToken: ch.botToken || '', allowFromText: ids.join(', ') }
    showTelegramModal.value = true
  }
}

async function fetchChannels() {
  loading.value = true
  try {
    const res = await getChannels()
    channels.value = res?.channels || []
    cache.channels = [...channels.value]
  } catch (e) { gm.error('获取频道失败: ' + (e.message || '')) }
  finally { loading.value = false }
}

async function saveWecom() {
  saving.value = true
  try {
    await saveChannel({ channelKey: 'wecom-app', enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true, corpId: wecomForm.value.corpId, corpSecret: wecomForm.value.corpSecret, agentId: wecomForm.value.agentId, token: wecomForm.value.token, encodingAESKey: wecomForm.value.encodingAESKey, dmPolicy: 'pairing', groupPolicy: 'open' })
    gm.success('频道配置已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function saveQQBot() {
  saving.value = true
  try {
    await saveChannel({ channelKey: 'qqbot', enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true, appId: qqForm.value.appId, clientSecret: qqForm.value.clientSecret })
    gm.success('QQ 机器人已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function saveDingTalk() {
  saving.value = true
  try {
    await saveChannel({ channelKey: 'dingtalk', enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true, clientId: dingtalkForm.value.clientId, clientSecret: dingtalkForm.value.clientSecret, robotCode: dingtalkForm.value.robotCode, corpId: dingtalkForm.value.corpId, agentId: dingtalkForm.value.agentId, dmPolicy: 'open', groupPolicy: 'open', debug: false, messageType: 'markdown' })
    gm.success('钉钉频道已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function saveFeishu() {
  saving.value = true
  try {
    await saveChannel({ channelKey: 'feishu', enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true, dmPolicy: 'open', accounts: { main: { appId: feishuForm.value.appId, appSecret: feishuForm.value.appSecret, botName: feishuForm.value.botName } } })
    gm.success('飞书频道已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function saveDiscord() {
  saving.value = true
  try {
    await saveChannel({
      channelKey: 'discord',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      token: discordForm.value.token,
      guildId: discordForm.value.guildId.trim(),
    })
    gm.success('Discord 频道已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function saveTelegram() {
  saving.value = true
  try {
    const allowIds = telegramForm.value.allowFromText.split(/[,，\n]/).map(s => s.trim()).filter(Boolean)
    const data = {
      channelKey: 'telegram',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      botToken: telegramForm.value.botToken,
      dmPolicy: 'allowlist',
      allowFrom: allowIds,
    }
    await saveChannel(data)
    gm.success('Telegram 频道已保存'); cancelForm(); await fetchChannels()
  } catch (e) { gm.error('保存失败: ' + (e.message || '')) }
  finally { saving.value = false }
}

async function deleteCh(ch) {
  const gmApi = gm.getGmApi()
  const doDelete = async () => {
    try { await deleteChannel({ channelKey: ch.key }); gm.success('频道已删除'); await fetchChannels() }
    catch (e) { gm.error('删除失败: ' + (e.message || '')) }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({ title: '删除频道', content: `确定删除「${getChannelDisplayName(ch.key)}」频道吗？`, positiveText: '确定', negativeText: '取消', onPositiveClick: doDelete })
  } else { if (confirm(`确定删除「${getChannelDisplayName(ch.key)}」？`)) doDelete() }
}

async function toggleCh(ch) {
  try { await toggleChannel({ channelKey: ch.key, enabled: !ch.enabled }); gm.success(ch.enabled ? '频道已禁用' : '频道已启用'); await fetchChannels() }
  catch (e) { gm.error('操作失败: ' + (e.message || '')) }
}

onMounted(() => {
  if (cache.channels !== null) { channels.value = [...cache.channels]; loading.value = false; return }
  fetchChannels()
})
</script>

<style scoped>
.channels-page { width: 100%; height: 100%; overflow-y: auto; padding: 20px 24px; }
.channels-container { max-width: 100%; margin: 0 auto; display: flex; flex-direction: column; gap: 20px; }

/* 双栏布局 */
.channel-body { display: grid; grid-template-columns: 1fr 220px; gap: 20px; }
.channel-main { display: flex; flex-direction: column; gap: 16px; }

/* 右侧监控面板 */
.channel-sidebar {
  display: flex; flex-direction: column; gap: 12px;
  position: sticky; top: 20px; align-self: start;
}
.sidebar-stat {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 20px 16px; border-radius: 14px;
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--jm-glass-border);
  box-shadow: 
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.03);
  transition: all 0.3s;
}
.sidebar-stat:hover { transform: translateY(-2px); box-shadow: 0 2px 4px rgba(0, 0, 0, 0.06), 0 8px 20px rgba(0, 0, 0, 0.05); }
.sidebar-stat-icon {
  width: 36px; height: 36px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
}
.sidebar-stat-icon.total { background: rgba(85, 105, 250, 0.1); color: #5569FA; }
.sidebar-stat-icon.connected { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
.sidebar-stat-icon.disconnected { background: rgba(156, 163, 175, 0.1); color: #9ca3af; }
.sidebar-stat-val { font-size: 26px; font-weight: 700; color: var(--jm-accent-7); line-height: 1.1; }
.sidebar-stat-label { font-size: 11px; color: var(--jm-accent-4); letter-spacing: 0.04em; }

/* 顶部 */
.channels-header { display: flex; align-items: flex-start; justify-content: space-between; }
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title { display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 600; color: var(--jm-accent-7); margin: 0; }
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }
.header-actions { display: flex; gap: 8px; align-items: center; }

.refresh-btn {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); cursor: pointer; transition: all 0.2s;
}
.refresh-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }
.refresh-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.spinning { animation: spin 0.8s linear infinite; }

.add-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px;
  border: none; background: var(--jm-primary-1); color: #fff;
  font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s;
}
.add-btn:hover { background: var(--jm-primary-2); }



/* 区块 */
.section-block {
  border: none; border-radius: 14px;
  background: var(--jm-glass-bg); padding: 18px;
  backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.03);
}
.section-title-row { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 14px; }
.section-label { margin: 0; font-size: 15px; font-weight: 600; color: var(--jm-accent-7); }
.section-hint { font-size: 12px; color: var(--jm-accent-4); }

/* 已接入频道小卡片 */
.connected-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; }
.connected-card {
  display: flex; flex-direction: column; justify-content: space-between;
  padding: 14px 16px; border-radius: 12px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  min-height: 90px;
  box-shadow: 
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.03);
}
.connected-card:hover {
  border-color: var(--jm-glass-border-hover);
  transform: translateY(-3px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.06), 0 12px 28px rgba(0, 0, 0, 0.06);
}
.cc-top { display: flex; align-items: center; gap: 10px; }
.cc-icon { color: var(--jm-primary-1); display: flex; align-items: center; flex-shrink: 0; }
.cc-icon svg { width: 22px; height: 22px; }
.cc-info { display: flex; flex-direction: column; min-width: 0; }
.cc-name { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cc-status { display: flex; align-items: center; gap: 4px; font-size: 11px; color: var(--jm-accent-4); }
.cc-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.cc-status.on .cc-dot { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.5); animation: led-breathe 2s ease-in-out infinite; }
.cc-status.off .cc-dot { background: var(--jm-accent-4); }
@keyframes led-breathe {
  0%, 100% { box-shadow: 0 0 4px rgba(34, 197, 94, 0.4); }
  50% { box-shadow: 0 0 10px rgba(34, 197, 94, 0.7), 0 0 20px rgba(34, 197, 94, 0.2); }
}
.cc-actions { display: flex; gap: 6px; margin-top: 10px; }
.cc-btn {
  padding: 3px 10px; border-radius: 4px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); font-size: 11px;
  cursor: pointer; transition: all 0.15s;
}
.cc-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-6); }
.cc-btn.danger:hover { border-color: var(--jm-error-color); color: var(--jm-error-color); }

/* 可用频道网格 */
.available-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; }
.available-card {
  display: flex; flex-direction: column; align-items: flex-start; gap: 6px;
  padding: 16px; border-radius: 12px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  color: var(--jm-accent-6);
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative; text-align: left;
  box-shadow: 
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.03);
}
.available-card:hover:not(.soon) {
  border-color: var(--jm-glass-border-hover);
  transform: translateY(-3px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.06), 0 12px 28px rgba(0, 0, 0, 0.06), 0 0 20px rgba(var(--jm-primary-1-rgb), 0.04);
}
.available-card.soon {
  opacity: 0.35; cursor: not-allowed;
  filter: blur(0.5px);
}
.available-card.active { border-color: rgba(34, 197, 94, 0.3); }
.avail-icon { display: flex; align-items: center; margin-bottom: 2px; }
.avail-name { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); }
.avail-auth { font-size: 11px; color: var(--jm-accent-4); }
.avail-badge {
  position: absolute; top: 8px; right: 8px;
  font-size: 10px; padding: 1px 6px; border-radius: 3px;
  background: rgba(var(--jm-accent-1-rgb), 0.5); color: var(--jm-accent-4);
}
.connected-badge { background: rgba(34, 197, 94, 0.1); color: #22c55e; }

/* 添加频道弹框 */
.modal-subtitle { font-size: 13px; color: var(--jm-accent-4); margin: 0 0 14px; }
.modal-type-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; }
.modal-type-card {
  display: flex; flex-direction: column; align-items: flex-start; gap: 6px;
  padding: 16px; border-radius: 10px; border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.3); cursor: pointer;
  transition: all 0.2s; text-align: left;
}
.modal-type-card:hover { border-color: var(--jm-primary-2); background: rgba(85, 105, 250, 0.06); }
.modal-type-icon { display: flex; color: var(--jm-accent-6); }
.modal-type-name { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); }
.modal-type-auth { font-size: 11px; color: var(--jm-accent-4); }

/* 配置表单 */
.channel-form { background: rgba(var(--jm-accent-1-rgb), 0.3); border: 1px solid var(--jm-glass-border); border-radius: 8px; padding: 18px 18px 6px; }
.modal-footer { display: flex; align-items: center; justify-content: space-between; }
.modal-actions { display: flex; gap: 8px; }
.help-link { font-size: 12px; color: var(--jm-primary-2); text-decoration: none; }
.help-link:hover { color: var(--jm-primary-1); text-decoration: underline; }
.telegram-hint {
  display: flex; align-items: center; gap: 6px;
  font-size: 12px; color: var(--jm-accent-5);
  padding: 8px 12px; margin-bottom: 12px;
  border-radius: 6px; background: rgba(var(--jm-primary-1-rgb), 0.05);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.1);
}
.env-hint { font-size: 11px; color: var(--jm-accent-4); }

/* 加载 */
.loading-state { display: flex; justify-content: center; padding: 40px; }
.loading-spinner { width: 24px; height: 24px; border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
