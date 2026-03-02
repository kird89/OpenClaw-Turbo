<template>
  <div class="agents-page">
    <div class="agents-container fade-in-up">
      <!-- 顶部操作栏 -->
      <div class="agents-header">
        <div class="header-left">
          <h2 class="page-title">
            <span v-html="icons.robot(20, 20)"></span>
            赛博员工
          </h2>
          <span class="header-hint">创建、编排你的赛博员工团队</span>
        </div>
        <div class="header-actions">
          <button class="mode-switch-btn" @click="viewMode = viewMode === 'list' ? 'cyber' : 'list'" :title="viewMode === 'list' ? '指挥中心' : '返回列表'">
            <svg v-if="viewMode === 'list'" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M8.766 8.766h6.469v6.469H8.766zm1.484 0V6.917m3.5 1.849V6.917m-3.5 10.166v-1.849m3.5 1.849v-1.849m1.484-4.984h1.849m-1.849 3.5h1.849m-10.165-3.5h1.848m-1.848 3.5h1.848" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M12 5.485a3.78 3.78 0 1 0-7.362 1.214a5.59 5.59 0 0 0 0 10.601A3.78 3.78 0 1 0 12 18.515m0-13.03a3.781 3.781 0 1 1 7.363 1.214a5.59 5.59 0 0 1 0 10.601A3.78 3.78 0 1 1 12 18.515" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            {{ viewMode === 'list' ? '指挥中心' : '返回列表' }}
          </button>
          <button class="refresh-btn" @click="fetchAgents()" :disabled="loading" title="刷新">
            <span :class="{ spinning: loading }" v-html="icons.refresh(16, 16)"></span>
          </button>

        </div>
      </div>

      <!-- ===== 指挥中心 (Cyber Topology) ===== -->
      <template v-if="viewMode === 'cyber'">
        <div class="topology-viewport" ref="cyberViewportRef">
          <div class="grid-bg"></div>
          <div class="scan-line"></div>
          <svg class="connections-layer" :viewBox="`0 0 ${cyberVpW} ${cyberVpH}`">
            <defs>
              <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stop-color="rgba(var(--jm-primary-1-rgb),0.1)" />
                <stop offset="50%" stop-color="rgba(var(--jm-primary-1-rgb),0.5)" />
                <stop offset="100%" stop-color="rgba(var(--jm-primary-1-rgb),0.1)" />
              </linearGradient>
              <filter id="glow"><feGaussianBlur stdDeviation="3" result="blur"/><feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
            </defs>
            <g v-for="conn in cyberConnections" :key="conn.id">
              <line :x1="conn.x1" :y1="conn.y1" :x2="conn.x2" :y2="conn.y2" stroke="url(#lineGrad)" stroke-width="1.5" stroke-dasharray="4 4" opacity="0.5" />
              <line :x1="conn.x1" :y1="conn.y1" :x2="conn.x2" :y2="conn.y2" stroke="url(#lineGrad)" stroke-width="1" filter="url(#glow)" opacity="0.3" />
              <circle r="3" fill="var(--jm-primary-1)" opacity="0.7" filter="url(#glow)">
                <animateMotion :dur="(3 + Math.random() * 2) + 's'" repeatCount="indefinite" :path="`M${conn.x1},${conn.y1} L${conn.x2},${conn.y2}`" />
              </circle>
              <circle r="3" fill="var(--jm-primary-1)" opacity="0.5">
                <animateMotion :dur="(3 + Math.random() * 2) + 's'" repeatCount="indefinite" :path="`M${conn.x2},${conn.y2} L${conn.x1},${conn.y1}`" />
              </circle>
            </g>
          </svg>
          <div v-for="node in cyberNodes" :key="node.id" class="agent-node" :class="[node.role, node.status]" :style="{ left: node.x + 'px', top: node.y + 'px' }" @click="viewMode = 'list'; openDetail(node)">
            <div class="pulse-ring" :class="node.status"></div>
            <div class="pulse-ring delay" :class="node.status"></div>
            <div class="node-core"><span class="node-icon" v-html="(icons[node.avatar] || icons.robot)(node.role === 'main' ? 28 : 22, node.role === 'main' ? 28 : 22)"></span></div>
            <span class="node-label">{{ node.name }}</span>
            <span class="node-status-text">{{ cyberStatusLabel(node.status) }}</span>
          </div>
          <div v-if="!loading && agents.length === 0" class="empty-state"><span>暂无 Agent，点击「新建 Agent」创建</span></div>
        </div>
        <div class="stats-bar" v-if="agents.length > 0">
          <div class="stat-item"><span class="stat-value">{{ agents.length }}</span><span class="stat-label">Agent 总数</span></div>
          <div class="stat-item"><span class="stat-dot idle"></span><span class="stat-value">{{ cyberCount('idle') }}</span><span class="stat-label">待命</span></div>
          <div class="stat-item"><span class="stat-dot thinking"></span><span class="stat-value">{{ cyberCount('thinking') }}</span><span class="stat-label">思考中</span></div>
          <div class="stat-item"><span class="stat-dot acting"></span><span class="stat-value">{{ cyberCount('acting') }}</span><span class="stat-label">执行中</span></div>
          <div class="stat-item"><span class="stat-dot error"></span><span class="stat-value">{{ cyberCount('error') }}</span><span class="stat-label">异常</span></div>
        </div>
      </template>

      <!-- ===== Agent 列表 / 编辑详情切换 ===== -->
      <template v-if="viewMode === 'list' && !editingAgent">
        <!-- 动态网格背景 -->
        <div class="cyber-grid-bg"></div>

        <!-- 局部工具栏 -->
        <div class="list-toolbar">
          <div class="search-box">
            <span class="search-icon" v-html="icons.search(14, 14)"></span>
            <input v-model="searchQuery" type="text" placeholder="搜索 Agent..." class="search-input" />
          </div>
          <div class="toolbar-actions">
            <button class="create-btn" @click="openCreateModal">
              <span v-html="icons.plus(14, 14)"></span>
              新建 Agent
            </button>
          </div>
        </div>

        <div class="list-layout" v-if="!loading">
          <!-- 左侧：Agent 卡片网格 -->
          <div class="card-grid">
            <div
              v-for="agent in filteredAgents"
              :key="agent.id"
              class="agent-card"
              :class="[agent.status, { 'is-main': agent.role === 'main' }]"
              @click="openDetail(agent)"
            >
              <!-- 卡片顶部光条 -->
              <div class="card-glow-bar" :class="agent.status"></div>

              <div class="card-header">
                <!-- 头像 + 光晕 -->
                <div class="card-avatar" :class="agent.status">
                  <div class="avatar-halo" :class="agent.status"></div>
                  <span class="avatar-icon" v-html="(icons[agent.avatar] || icons.robot)(24, 24)"></span>
                </div>
                <div class="card-title-area">
                  <div class="card-name">
                    {{ agent.name }}
                    <span class="role-tag" :class="agent.role">{{ agent.role === 'main' ? '主控' : '专员' }}</span>
                  </div>
                  <div class="card-model" v-if="agent.model">{{ agent.model }}</div>
                </div>
                <!-- 操作按钮 -->
                <div class="card-actions" @click.stop>
                  <button class="act-btn" @click="openEditModal(agent)" title="编辑">
                    <span v-html="icons.pencil(13, 13)"></span>
                  </button>
                  <button v-if="agent.role !== 'main'" class="act-btn danger" @click="confirmDelete(agent)" :disabled="deleting === agent.id" title="删除">
                    <span v-if="deleting === agent.id" class="spin" v-html="icons.loader(13, 13)"></span>
                    <span v-else v-html="icons.trash(13, 13)"></span>
                  </button>
                </div>
              </div>

              <div class="card-desc">{{ agent.description || '暂无描述' }}</div>

              <!-- 状态标签 -->
              <div class="card-status-row">
                <span class="status-badge" :class="agent.status">
                  <span class="status-dot-mini" :class="agent.status"></span>
                  {{ statusLabel(agent.status) }}
                </span>
              </div>

            </div>

            <!-- 空状态 -->
            <div v-if="filteredAgents.length === 0 && agents.length > 0" class="empty-search">
              未找到匹配的 Agent
            </div>
          </div>

          <!-- 右侧：团队概览面板 -->
          <div class="team-panel">
            <div class="panel-section">
              <div class="panel-title">团队状态</div>
              <!-- 环形进度 -->
              <div class="ring-chart-wrap">
                <svg class="ring-chart" viewBox="0 0 100 100">
                  <circle cx="50" cy="50" r="40" fill="none" stroke="var(--jm-accent-2)" stroke-width="8" />
                  <circle cx="50" cy="50" r="40" fill="none" stroke="var(--jm-primary-1)" stroke-width="8"
                    :stroke-dasharray="onlineRingDash" stroke-dashoffset="0"
                    stroke-linecap="round" transform="rotate(-90 50 50)"
                    style="transition: stroke-dasharray 0.6s ease" />
                </svg>
                <div class="ring-label">
                  <span class="ring-val">{{ onlineCount }}</span>
                  <span class="ring-sub">/ {{ agents.length }}</span>
                </div>
              </div>
              <div class="ring-hint">在线员工</div>
            </div>

            <div class="panel-section">
              <div class="panel-title">团队组成</div>
              <div class="team-comp">
                <div class="comp-row">
                  <span class="comp-dot main"></span>
                  <span class="comp-label">主控</span>
                  <span class="comp-val">{{ agents.filter(a => a.role === 'main').length }}</span>
                </div>
                <div class="comp-row">
                  <span class="comp-dot specialist"></span>
                  <span class="comp-label">专员</span>
                  <span class="comp-val">{{ agents.filter(a => a.role !== 'main').length }}</span>
                </div>
                <div class="comp-row">
                  <span class="comp-dot idle"></span>
                  <span class="comp-label">待命</span>
                  <span class="comp-val">{{ cyberCount('idle') }}</span>
                </div>
                <div class="comp-row">
                  <span class="comp-dot thinking"></span>
                  <span class="comp-label">思考中</span>
                  <span class="comp-val">{{ cyberCount('thinking') }}</span>
                </div>
                <div class="comp-row">
                  <span class="comp-dot acting"></span>
                  <span class="comp-label">执行中</span>
                  <span class="comp-val">{{ cyberCount('acting') }}</span>
                </div>
                <div class="comp-row">
                  <span class="comp-dot error"></span>
                  <span class="comp-label">异常</span>
                  <span class="comp-val">{{ cyberCount('error') }}</span>
                </div>
              </div>
            </div>


          </div>
        </div>

        <div v-else class="loading-state">
          <div class="loading-spinner"></div>
          <span>加载中...</span>
        </div>
      </template>

      <!-- Agent 详情编辑 -->
      <template v-if="viewMode === 'list' && editingAgent">
        <div class="detail-header">
          <button class="back-btn" @click="exitDetail">
            <span v-html="icons.arrowLeft(16, 16)"></span>
            返回列表
          </button>
          <div class="detail-info">
            <span class="detail-icon" v-html="icons.robot(20, 20)"></span>
            <span class="detail-name">{{ editingAgent.name }}</span>
            <span class="role-tag" :class="editingAgent.role">
              {{ editingAgent.role === 'main' ? '主控' : '专员' }}
            </span>
          </div>
          <div class="header-actions">
            <!-- 主控: 管理风格下拉 -->
            <n-dropdown v-if="editingAgent.role === 'main'" :options="presetDropdownOptions" trigger="click" @select="onPresetSelect" :to="false">
              <button class="tpl-btn">
                <span v-html="icons.sparkles(14, 14)"></span>
                管理风格
                <span v-html="icons.chevronDown(12, 12)"></span>
              </button>
            </n-dropdown>
            <!-- 专员: 模板下拉 -->
            <n-dropdown v-else :options="templateDropdownOptions" trigger="click" @select="onTemplateSelect" :to="false">
              <button class="tpl-btn">
                <span v-html="icons.star(14, 14)"></span>
                模板
                <span v-html="icons.chevronDown(12, 12)"></span>
              </button>
            </n-dropdown>
          </div>
        </div>

        <!-- 灵魂注入标题 -->
        <div class="soul-inject-title">
          <span class="soul-glow"></span>
          赋予你的 Agent 灵魂
        </div>

        <!-- 模式切换 -->
        <div class="mode-toggle">
          <button :class="{ active: editMode === 'visual' }" @click="editMode = 'visual'">
            <span v-html="icons.user(14, 14)"></span> 可视化建模
          </button>
          <button :class="{ active: editMode === 'code' }" @click="editMode = 'code'">
            <span v-html="icons.brainChip(14, 14)"></span> 专家模式
          </button>
        </div>

        <!-- ======= 可视化建模 ======= -->
        <template v-if="editMode === 'visual' && !detailLoading">
          <!-- Identity 职能模块 -->
          <div class="persona-section" :class="{ active: activeTab === 'IDENTITY' }">
            <div class="section-header" @click="toggleSection('IDENTITY')">
              <span class="section-icon" v-html="icons.fingerprint(16, 16)"></span>
              <span class="section-title">身份设定</span>
              <span class="section-hint">我是谁 · Identity</span>
              <span v-if="isModified('IDENTITY')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.IDENTITY }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.IDENTITY">
              <div class="field-row">
                <label>名称</label>
                <n-input v-model:value="persona.identity.name" placeholder="起个名字..." size="small" />
              </div>
              <div class="field-row">
                <label>身份类型</label>
                <n-select v-model:value="persona.identity.creature" :options="creatureOptions" size="small" />
              </div>

              <div class="field-row">
                <label>氛围</label>
                <div class="tag-group">
                  <button
                    v-for="v in vibeOptions"
                    :key="v.value"
                    class="vibe-tag"
                    :class="{ selected: persona.identity.vibes.includes(v.value), [v.glow]: persona.identity.vibes.includes(v.value) }"
                    @click="toggleVibe(v.value)"
                  >{{ v.label }}</button>
                </div>
              </div>
            </div>
          </div>

          <!-- User 服务对象 -->
          <div class="persona-section" :class="{ active: activeTab === 'USER' }">
            <div class="section-header" @click="toggleSection('USER')">
              <span class="section-icon" v-html="icons.users(16, 16)"></span>
              <span class="section-title">服务对象</span>
              <span class="section-hint">你是谁 · User</span>
              <span v-if="isModified('USER')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.USER }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.USER">
              <div class="field-row">
                <label>用户名</label>
                <n-input v-model:value="persona.user.name" placeholder="你的名字" size="small" />
              </div>
              <div class="field-row">
                <label>如何称呼</label>
                <n-input v-model:value="persona.user.callName" placeholder="Boss / 老板 / 大佬..." size="small" />
              </div>
              <div class="field-row">
                <label>背景</label>
                <n-input v-model:value="persona.user.context" type="textarea" placeholder="用户关心什么？在做什么项目？有什么偏好？" :rows="2" size="small" />
              </div>
            </div>
          </div>

          <!-- Soul 共鸣模块 -->
          <div class="persona-section" :class="{ active: activeTab === 'SOUL' }">
            <div class="section-header" @click="toggleSection('SOUL')">
              <span class="section-icon" v-html="icons.sparkles(16, 16)"></span>
              <span class="section-title">性格灵魂</span>
              <span class="section-hint">怎么聊 · Soul</span>
              <span v-if="isModified('SOUL')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.SOUL }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.SOUL">
              <div class="slider-row">
                <label>逻辑严谨度</label>
                <div class="slider-labels"><span>随意</span><span>严谨</span></div>
                <n-slider v-model:value="persona.soul.formality" :min="0" :max="100" :step="1" />
                <div class="slider-desc">{{ sliderDesc('formality', persona.soul.formality) }}</div>
              </div>
              <div class="slider-row">
                <label>输出熵值</label>
                <div class="slider-labels"><span>精简</span><span>详细</span></div>
                <n-slider v-model:value="persona.soul.verbosity" :min="0" :max="100" :step="1" />
                <div class="slider-desc">{{ sliderDesc('verbosity', persona.soul.verbosity) }}</div>
              </div>
              <div class="slider-row">
                <label>自主决策力</label>
                <div class="slider-labels"><span>被动</span><span>主动</span></div>
                <n-slider v-model:value="persona.soul.initiative" :min="0" :max="100" :step="1" />
                <div class="slider-desc">{{ sliderDesc('initiative', persona.soul.initiative) }}</div>
              </div>
              <div class="slider-row">
                <label>共情指数</label>
                <div class="slider-labels"><span>理性</span><span>感性</span></div>
                <n-slider v-model:value="persona.soul.empathy" :min="0" :max="100" :step="1" />
                <div class="slider-desc">{{ sliderDesc('empathy', persona.soul.empathy) }}</div>
              </div>
              <div class="field-row" style="margin-top: 8px;">
                <label>核心原则</label>
                <div class="tag-group">
                  <button
                    v-for="p in principleOptions"
                    :key="p"
                    class="vibe-tag"
                    :class="{ selected: persona.soul.principles.includes(p) }"
                    @click="togglePrinciple(p)"
                  >{{ p }}</button>
                </div>
              </div>
            </div>
          </div>

          <!-- Tools 工具箱 -->
          <div class="persona-section" :class="{ active: activeTab === 'TOOLS' }">
            <div class="section-header" @click="toggleSection('TOOLS')">
              <span class="section-icon" v-html="icons.tool(16, 16)"></span>
              <span class="section-title">工具箱</span>
              <span class="section-hint">使用偏好 · Tools</span>
              <span v-if="isModified('TOOLS')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.TOOLS }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.TOOLS">
              <div class="field-row">
                <label>工具使用倾向</label>
                <div class="tag-group">
                  <button class="vibe-tag" :class="{ selected: persona.tools.tendency === 'aggressive' }" @click="persona.tools.tendency = 'aggressive'; syncPersonaToMarkdown('TOOLS')">⚡ 激进（一言不合就执行）</button>
                  <button class="vibe-tag" :class="{ selected: persona.tools.tendency === 'balanced' }" @click="persona.tools.tendency = 'balanced'; syncPersonaToMarkdown('TOOLS')">⚖️ 平衡（视情况而定）</button>
                  <button class="vibe-tag" :class="{ selected: persona.tools.tendency === 'conservative' }" @click="persona.tools.tendency = 'conservative'; syncPersonaToMarkdown('TOOLS')">🛡️ 保守（先问再执行）</button>
                </div>
              </div>
              <div class="field-row">
                <label>浏览器约束</label>
                <n-input v-model:value="persona.tools.constraints.browser" type="textarea" placeholder="例如：优先搜索英文资料，避免访问广告站点..." :rows="2" size="small" @input="syncPersonaToMarkdown('TOOLS')" />
              </div>
              <div class="field-row">
                <label>终端执行约束</label>
                <n-input v-model:value="persona.tools.constraints.exec" type="textarea" placeholder="例如：执行前必须先输出完整命令供确认..." :rows="2" size="small" @input="syncPersonaToMarkdown('TOOLS')" />
              </div>
              <div class="field-row">
                <label>写文件约束</label>
                <n-input v-model:value="persona.tools.constraints.write" type="textarea" placeholder="例如：写入前先备份原文件..." :rows="2" size="small" @input="syncPersonaToMarkdown('TOOLS')" />
              </div>
            </div>
          </div>

          <!-- Bootstrap 入职 -->
          <div class="persona-section" :class="{ active: activeTab === 'BOOTSTRAP' }">
            <div class="section-header" @click="toggleSection('BOOTSTRAP')">
              <span class="section-icon" v-html="icons.rocketSimple(16, 16)"></span>
              <span class="section-title">入职自检</span>
              <span class="section-hint">首次启动 · Bootstrap</span>
              <span v-if="isModified('BOOTSTRAP')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.BOOTSTRAP }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.BOOTSTRAP">
              <div class="field-row">
                <label>初始化动作</label>
                <div class="check-group">
                  <label class="check-item"><input type="checkbox" v-model="persona.bootstrap.greeting" @change="syncPersonaToMarkdown('BOOTSTRAP')" /> 启动时主动打招呼</label>
                  <label class="check-item"><input type="checkbox" v-model="persona.bootstrap.scanProject" @change="syncPersonaToMarkdown('BOOTSTRAP')" /> 启动时扫描工作区目录结构</label>
                </div>
              </div>
              <div class="field-row">
                <label>首次唤醒任务</label>
                <n-input v-model:value="persona.bootstrap.firstTask" type="textarea" placeholder="新 Agent 启动后必须完成的第一个任务，例如：去 GitHub 拉取 XXX 仓库并分析依赖..." :rows="3" size="small" @input="syncPersonaToMarkdown('BOOTSTRAP')" />
              </div>
            </div>
          </div>

          <!-- Heartbeat 心跳 -->
          <div class="persona-section" :class="{ active: activeTab === 'HEARTBEAT' }">
            <div class="section-header" @click="toggleSection('HEARTBEAT')">
              <span class="section-icon" v-html="icons.activity(16, 16)"></span>
              <span class="section-title">赛博心跳</span>
              <span class="section-hint">闲置思考 · Heartbeat</span>
              <span v-if="isModified('HEARTBEAT')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.HEARTBEAT }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.HEARTBEAT">
              <div class="field-row">
                <label>被唤醒时的主要目标</label>
                <div class="tag-group">
                  <button class="vibe-tag" :class="{ selected: persona.heartbeat.wakeGoal === 'summarize' }" @click="persona.heartbeat.wakeGoal = 'summarize'; syncPersonaToMarkdown('HEARTBEAT')">📋 整理对话摘要</button>
                  <button class="vibe-tag" :class="{ selected: persona.heartbeat.wakeGoal === 'todo' }" @click="persona.heartbeat.wakeGoal = 'todo'; syncPersonaToMarkdown('HEARTBEAT')">✅ 检查待办事项</button>
                  <button class="vibe-tag" :class="{ selected: persona.heartbeat.wakeGoal === 'crawl' }" @click="persona.heartbeat.wakeGoal = 'crawl'; syncPersonaToMarkdown('HEARTBEAT')">🌐 爬取特定资讯</button>
                  <button class="vibe-tag" :class="{ selected: persona.heartbeat.wakeGoal === 'standby' }" @click="persona.heartbeat.wakeGoal = 'standby'; syncPersonaToMarkdown('HEARTBEAT')">🔇 安静待机</button>
                </div>
              </div>
              <div class="slider-row">
                <label>主动打扰阈值</label>
                <div class="slider-labels"><span>极度克制</span><span>极度主动</span></div>
                <n-slider v-model:value="persona.heartbeat.disturbance" :min="0" :max="100" :step="1" />
                <div class="slider-desc">{{ persona.heartbeat.disturbance > 70 ? '一有风吹草动立刻发消息' : persona.heartbeat.disturbance > 30 ? '重要事项才会打扰' : '没重要的事也不发消息' }}</div>
              </div>
            </div>
          </div>

          <!-- Conventions 规范 -->
          <div class="persona-section" :class="{ active: activeTab === 'CONVENTIONS' }">
            <div class="section-header" @click="toggleSection('CONVENTIONS')">
              <span class="section-icon" v-html="icons.fileCode(16, 16)"></span>
              <span class="section-title">输出规范</span>
              <span class="section-hint">格式与风格 · Conventions</span>
              <span v-if="isModified('CONVENTIONS')" class="section-dot"></span>
              <span class="section-chevron" :class="{ open: !collapsed.CONVENTIONS }" v-html="icons.chevronDown(14, 14)"></span>
            </div>
            <div class="section-body" v-show="!collapsed.CONVENTIONS">
              <div class="field-row">
                <label>语言与框架偏好</label>
                <n-select v-model:value="persona.conventions.languages" :options="langOptions" multiple placeholder="点击添加标签..." size="small" @update:value="syncPersonaToMarkdown('CONVENTIONS')" />
              </div>
              <div class="field-row">
                <label>代码风格规范</label>
                <n-select v-model:value="persona.conventions.codeStyle" :options="codeStyleOptions" placeholder="选择代码规范" size="small" clearable @update:value="syncPersonaToMarkdown('CONVENTIONS')" />
              </div>
              <div class="field-row">
                <label>内容排版要求</label>
                <div class="check-group">
                  <label class="check-item"><input type="checkbox" v-model="persona.conventions.useMarkdown" @change="syncPersonaToMarkdown('CONVENTIONS')" /> 回复必须使用 Markdown 排版</label>
                  <label class="check-item"><input type="checkbox" v-model="persona.conventions.showThinking" @change="syncPersonaToMarkdown('CONVENTIONS')" /> 结论前展示思考过程</label>
                  <label class="check-item"><input type="checkbox" v-model="persona.conventions.useMermaid" @change="syncPersonaToMarkdown('CONVENTIONS')" /> 复杂逻辑附带 Mermaid 流程图</label>
                </div>
              </div>
            </div>
          </div>

          <!-- 保存栏 -->
          <div class="save-bar">
            <div class="save-bar-left">
              <span v-if="anyModified" class="modified-hint">● 有未保存的修改</span>
            </div>
            <div class="save-bar-right">
              <button class="tool-btn" @click="showPreview = true" title="预览生成的指令">
                <span v-html="icons.adjustments(14, 14)"></span> 预览指令
              </button>
              <button class="tool-btn" @click="resetFile" :disabled="saving">
                <span v-html="icons.refresh(14, 14)"></span> 重置
              </button>
              <button class="tool-btn save-btn" @click="saveAllVisual" :disabled="saving || !anyModified">
                <span v-if="saving" class="spin" v-html="icons.loader(14, 14)"></span>
                <span v-else v-html="icons.save(14, 14)"></span>
                保存人格
              </button>
            </div>
          </div>

          <!-- 预览指令浮层 -->
          <div v-if="showPreview" class="preview-overlay" @click.self="showPreview = false">
            <div class="preview-panel">
              <div class="preview-header">
                <span>指令预览</span>
                <button class="preview-close" @click="showPreview = false">✕</button>
              </div>
              <div class="preview-tabs">
                <button v-for="tab in filteredTabs" :key="tab.key" :class="{ active: previewTab === tab.key }" @click="previewTab = tab.key">{{ tab.desc }}</button>
              </div>
              <pre class="preview-content">{{ files[previewTab] }}</pre>
            </div>
          </div>
        </template>

        <!-- ======= 专家模式 (代码) ======= -->
        <template v-else-if="editMode === 'code' && !detailLoading">
          <div class="tab-bar">
            <button v-for="tab in filteredTabs" :key="tab.key" class="tab-item" :class="{ active: activeTab === tab.key }" @click="switchTab(tab.key)">
              <span class="tab-icon" v-html="tab.icon"></span>
              <span class="tab-label">{{ tab.label }}</span>
              <span class="tab-desc">{{ tab.desc }}</span>
              <span v-if="isModified(tab.key)" class="tab-dot"></span>
            </button>
          </div>
          <div class="editor-panel">
            <div class="editor-toolbar">
              <div class="toolbar-left">
                <span class="file-badge">{{ activeFileName }}</span>
                <span v-if="isModified(activeTab)" class="modified-hint">● 已修改</span>
              </div>
              <div class="toolbar-right">
                <button class="tool-btn" @click="resetFile" :disabled="saving">
                  <span v-html="icons.refresh(14, 14)"></span> 恢复默认
                </button>
                <button class="tool-btn save-btn" @click="saveFile" :disabled="saving || !isModified(activeTab)">
                  <span v-if="saving" class="spin" v-html="icons.loader(14, 14)"></span>
                  <span v-else v-html="icons.save(14, 14)"></span> 保存
                </button>
              </div>
            </div>
            <div class="editor-wrapper">
              <textarea ref="editorRef" class="editor-textarea" :value="currentContent" @input="onInput" placeholder="在此编辑 Markdown 内容..." spellcheck="false"></textarea>
            </div>
          </div>
        </template>

        <div v-if="detailLoading" class="editor-loading">
          <div class="loading-spinner"></div>
          <span>加载中...</span>
        </div>
      </template>
    </div>

    <!-- 创建/编辑 Agent Modal -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="modalTitle"
      :style="{ width: '480px' }"
      :mask-closable="false"
      :bordered="false"
    >
      <n-form ref="formRef" :model="formData" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="支持中英文、数字、下划线、短横线" :allow-input="v => /^[\u4e00-\u9fa5a-zA-Z0-9_-]*$/.test(v)" />
        </n-form-item>
        <n-form-item label="头像" path="avatar">
          <div class="modal-avatar-grid">
            <div 
              v-for="opt in avatarOptions" 
              :key="opt" 
              class="modal-avatar-option" 
              :class="{ active: formData.avatar === opt }"
              @click="formData.avatar = opt"
              v-html="icons[opt](20, 20)"
            ></div>
          </div>
        </n-form-item>
        <n-form-item label="角色" path="role">
          <n-input :value="formData.role === 'main' ? '主控 (Main)' : '专员 (Specialist)'" disabled />
        </n-form-item>
        <n-form-item label="从属于" path="parentId" v-if="formData.role === 'specialist'">
          <n-select v-model:value="formData.parentId" :options="parentOptions" placeholder="选择主 Agent" clearable />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="这个 Agent 负责什么？" :rows="2" />
        </n-form-item>
        <n-form-item label="对话模型" path="model">
          <n-select v-model:value="formData.model" :options="modelOptions" placeholder="选择 AI 模型" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-footer">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" @click="submitForm" :loading="submitting">
            {{ formData.id ? '保存修改' : '创建 Agent' }}
          </n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, reactive, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { NDropdown, NModal, NForm, NFormItem, NInput, NSelect, NButton, NSlider } from 'naive-ui'
import {
  listAgents, createAgent, updateAgent, deleteAgent, getAgentDetail,
  getAgentTemplates, applyAgentTemplate, saveAgentFile, resetAgentFile,
  getConfiguredModels
} from '@/api/agent'
import gm from '@/utils/gmssh'
import { icons, avatarOptions } from '@/components/icons'
import { personaPresets } from '@/components/persona-presets'

const loading = ref(true)
const saving = ref(false)
const detailLoading = ref(false)
const submitting = ref(false)
const deleting = ref(null)
const agents = ref([])
const templates = ref([])
const editMode = ref('visual')
const initializing = ref(false)
const showPreview = ref(false)
const previewTab = ref('IDENTITY')
const modelOptions = ref([])
const defaultModel = ref('')
const viewMode = ref('list') // 'list' | 'cyber'
const searchQuery = ref('')

// ===== 搜索过滤 =====
const filteredAgents = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return agents.value
  return agents.value.filter(a =>
    a.name.toLowerCase().includes(q) ||
    (a.description || '').toLowerCase().includes(q) ||
    (a.model || '').toLowerCase().includes(q)
  )
})

// ===== 模拟统计（基于 agent name hash 生成稳定数据）=====
function simpleHash(str) {
  let h = 0
  for (let i = 0; i < str.length; i++) h = ((h << 5) - h + str.charCodeAt(i)) | 0
  return Math.abs(h)
}
function agentMockStats(agent) {
  const h = simpleHash(agent.id || agent.name)
  const tasks = 10 + (h % 200)
  const hours = 1 + (h % 72)
  const uptime = hours >= 24 ? Math.floor(hours / 24) + 'd' : hours + 'h'
  const tokens = ((h % 50) + 5) + 'k'
  return { tasks, uptime, tokens }
}

// ===== 能力标签 =====
const capabilityDefs = [
  { key: 'code', label: '代码', icon: icons.code(13, 13), keywords: ['代码', 'code', '开发', '编程', 'dev'] },
  { key: 'terminal', label: '终端', icon: icons.terminal(13, 13), keywords: ['终端', 'terminal', 'shell', '命令'] },
  { key: 'search', label: '搜索', icon: icons.search(13, 13), keywords: ['搜索', 'search', '查找', '检索'] },
  { key: 'globe', label: '网络', icon: icons.globe(13, 13), keywords: ['网络', 'web', '浏览', 'browse', 'http'] },
  { key: 'database', label: '数据', icon: icons.database(13, 13), keywords: ['数据', 'data', 'sql', '数据库'] },
  { key: 'palette', label: '设计', icon: icons.palette(13, 13), keywords: ['设计', 'design', 'ui', '界面', '绘图'] },
  { key: 'cpu', label: '算力', icon: icons.cpu(13, 13), keywords: ['算力', 'compute', 'gpu', 'cpu', '运算'] },
]
function agentCapabilities(agent) {
  const desc = ((agent.description || '') + ' ' + (agent.name || '')).toLowerCase()
  const matched = capabilityDefs.filter(c => c.keywords.some(kw => desc.includes(kw)))
  // 主控默认给全能力，专员至少给 2 个
  if (agent.role === 'main') return capabilityDefs.slice(0, 4)
  if (matched.length > 0) return matched.slice(0, 4)
  const h = simpleHash(agent.id || agent.name)
  return [capabilityDefs[h % capabilityDefs.length], capabilityDefs[(h + 2) % capabilityDefs.length]]
}

// ===== 团队概览 =====
const onlineCount = computed(() => agents.value.filter(a => a.status !== 'error').length)
const onlineRingDash = computed(() => {
  const total = agents.value.length || 1
  const ratio = onlineCount.value / total
  const circ = 2 * Math.PI * 40 // ~251.33
  return `${circ * ratio} ${circ * (1 - ratio)}`
})

// ===== 活动流 =====
const activityFeed = computed(() => {
  const feed = []
  const now = new Date()
  const statusVerbs = {
    idle: '进入待命状态',
    thinking: '正在分析上下文...',
    acting: '正在执行任务',
    error: '报告了一个异常',
  }
  agents.value.forEach((agent, i) => {
    const mins = 2 + i * 7
    const t = new Date(now - mins * 60000)
    const timeStr = String(t.getHours()).padStart(2, '0') + ':' + String(t.getMinutes()).padStart(2, '0')
    feed.push({
      time: timeStr,
      type: agent.status || 'idle',
      text: `${agent.name} ${statusVerbs[agent.status] || '已上线'}`,
    })
  })
  // Add some synthetic entries
  if (agents.value.length > 0) {
    const a0 = agents.value[0]
    const t1 = new Date(now - 15 * 60000)
    feed.push({
      time: String(t1.getHours()).padStart(2, '0') + ':' + String(t1.getMinutes()).padStart(2, '0'),
      type: 'idle',
      text: `${a0.name} 完成了任务分发`,
    })
    const t2 = new Date(now - 30 * 60000)
    feed.push({
      time: String(t2.getHours()).padStart(2, '0') + ':' + String(t2.getMinutes()).padStart(2, '0'),
      type: 'thinking',
      text: `系统健康检查完成`,
    })
  }
  return feed.sort((a, b) => b.time.localeCompare(a.time)).slice(0, 8)
})

// ===== 指挥中心 (Cyber Topology) =====
const cyberViewportRef = ref(null)
const cyberVpW = ref(800)
const cyberVpH = ref(500)
let cyberResizeObs = null

function cyberStatusLabel(status) {
  const map = { idle: 'IDLE', thinking: 'THINKING...', acting: 'EXECUTING', error: 'ERROR' }
  return map[status] || (status || '').toUpperCase()
}

function cyberCount(status) {
  return agents.value.filter(a => a.status === status).length
}

const cyberNodes = computed(() => {
  const list = agents.value
  if (list.length === 0) return []
  const w = cyberVpW.value, h = cyberVpH.value
  const cx = w / 2, cy = h / 2
  const mainAgent = list.find(a => a.role === 'main')
  const specialists = list.filter(a => a.role !== 'main')
  const nodes = []
  if (mainAgent) nodes.push({ ...mainAgent, x: cx - 40, y: cy - 50 })
  const radius = Math.min(w, h) * 0.32
  specialists.forEach((agent, i) => {
    const angle = (2 * Math.PI * i) / Math.max(specialists.length, 1) - Math.PI / 2
    nodes.push({ ...agent, x: cx + radius * Math.cos(angle) - 35, y: cy + radius * Math.sin(angle) - 35 })
  })
  return nodes
})

const cyberConnections = computed(() => {
  const main = cyberNodes.value.find(n => n.role === 'main')
  if (!main) return []
  return cyberNodes.value.filter(n => n.role !== 'main').map(n => ({
    id: `${main.id}-${n.id}`, x1: main.x + 40, y1: main.y + 40, x2: n.x + 35, y2: n.y + 35,
  }))
})

function updateCyberViewport() {
  if (cyberViewportRef.value) {
    cyberVpW.value = cyberViewportRef.value.clientWidth
    cyberVpH.value = cyberViewportRef.value.clientHeight
  }
}

watch(viewMode, async (mode) => {
  if (mode === 'cyber') {
    await nextTick()
    updateCyberViewport()
    if (cyberViewportRef.value) {
      cyberResizeObs = new ResizeObserver(updateCyberViewport)
      cyberResizeObs.observe(cyberViewportRef.value)
    }
  } else {
    if (cyberResizeObs) { cyberResizeObs.disconnect(); cyberResizeObs = null }
  }
})

onUnmounted(() => {
  if (cyberResizeObs) cyberResizeObs.disconnect()
})

// Slider dynamic descriptions
const sliderDescMap = {
  formality: [
    [30, '轻松随意，像朋友聊天一样自然'],
    [70, '灵活平衡，根据场景自适应'],
    [101, '专业严谨，每个输出都经过严格校验'],
  ],
  verbosity: [
    [30, '极度精简，只说最关键的'],
    [70, '简洁与详尽并重，该展开时展开'],
    [101, '丰富详尽，提供完整的上下文和解释'],
  ],
  initiative: [
    [30, '待命模式，等待明确指令后行动'],
    [70, '适度主动，发现问题会提醒'],
    [101, '积极探索，主动发现问题并提出方案'],
  ],
  empathy: [
    [30, '纯理性分析，就事论事不带感情'],
    [70, '理性与共情并重，能理解用户情绪'],
    [101, '高度共情，细腥感受并关心用户感受'],
  ]
}
function sliderDesc(key, val) {
  const tiers = sliderDescMap[key]
  for (const [threshold, desc] of tiers) {
    if (val < threshold) return desc
  }
  return tiers[tiers.length - 1][1]
}

// ========== 人格预设 ==========
const activePreset = ref(null)

function applyPersonaPreset(preset) {
  activePreset.value = preset.key
  files.value.IDENTITY = preset.identity
  files.value.USER = preset.user
  files.value.SOUL = preset.soul
  parseMarkdownToPersona()
  setTimeout(() => syncPersonaToMarkdown('all'), 10)
}

const presetDropdownOptions = computed(() =>
  personaPresets.map(p => ({
    key: p.key,
    label: p.name + ' — ' + p.desc,
  }))
)

function onPresetSelect(key) {
  const preset = personaPresets.find(p => p.key === key)
  if (preset) applyPersonaPreset(preset)
}

// ========== 人格建模数据 ==========
const persona = reactive({
  identity: { name: '', creature: '', vibes: [], avatar: 'brain' },
  user: { name: '', callName: '', timezone: '', context: '' },
  soul: { formality: 50, verbosity: 50, initiative: 50, empathy: 50, principles: [] },
  tools: { tendency: 'balanced', constraints: {} },
  bootstrap: { greeting: true, scanProject: false, firstTask: '' },
  heartbeat: { wakeGoal: 'summarize', disturbance: 30 },
  conventions: { languages: [], codeStyle: '', useMarkdown: true, showThinking: false, useMermaid: false },
})

// 卡片折叠状态
const collapsed = reactive({
  IDENTITY: true, USER: true, SOUL: true,
  TOOLS: true, BOOTSTRAP: true, HEARTBEAT: true, CONVENTIONS: true,
})
function toggleSection(key) { collapsed[key] = !collapsed[key] }

const creatureOptions = [
  { label: '🤖 AI 助手', value: 'AI助手' },
  { label: '💻 数字生命', value: '数字生命' },
  { label: '🌟 资深专家', value: '资深专家' },
  { label: '👻 虚拟助理', value: '虚拟助理' },
  { label: '😺 猫娘', value: '猫娘' },
  { label: '🔥 故障猎人', value: '故障猎人' },
]

const vibeOptions = [
  { label: '❄️ 冷静', value: '冷静', glow: 'glow-blue' },
  { label: '🔥 热情', value: '热情', glow: 'glow-orange' },
  { label: '🎯 犯利', value: '犯利', glow: 'glow-green' },
  { label: '💡 智慧', value: '智慧', glow: 'glow-purple' },
  { label: '🫡 温暖', value: '温暖', glow: 'glow-yellow' },
  { label: '⚡ 高效', value: '高效', glow: 'glow-cyan' },
  { label: '🌊 深沉', value: '深沉', glow: 'glow-blue' },
  { label: '🎭 幽默', value: '幽默', glow: 'glow-pink' },
]

const principleOptions = ['先做再问', '代码优先', '有观点', '谨慎操作', '结构化输出', '简洁明了', '自我辭代', '主动探索']

function toggleVibe(v) {
  const idx = persona.identity.vibes.indexOf(v)
  if (idx >= 0) persona.identity.vibes.splice(idx, 1)
  else persona.identity.vibes.push(v)
  syncPersonaToMarkdown('IDENTITY')
}

function togglePrinciple(p) {
  const idx = persona.soul.principles.indexOf(p)
  if (idx >= 0) persona.soul.principles.splice(idx, 1)
  else persona.soul.principles.push(p)
  syncPersonaToMarkdown('SOUL')
}

const langOptions = [
  'Python', 'TypeScript', 'JavaScript', 'Go', 'Rust', 'Java', 'C++',
  'React', 'Vue', 'Next.js', 'Tailwind', 'Node.js', 'Docker', 'Kubernetes'
].map(v => ({ label: v, value: v }))

const codeStyleOptions = [
  { label: 'PEP8 (Python)', value: 'PEP8' },
  { label: 'Google Style Guide', value: 'Google' },
  { label: '遵循项目 ESLint 配置', value: 'ESLint' },
  { label: 'Standard.js', value: 'Standard' },
  { label: 'Airbnb Style', value: 'Airbnb' },
]

// ===== Markdown ↔ Persona 双向转换 =====
function syncPersonaToMarkdown(section) {
  if (initializing.value) return
  if (section === 'IDENTITY' || section === 'all') {
    const p = persona.identity
    files.value.IDENTITY = `# 身份设定

- **名称:** ${p.name || '未命名'}
- **类型:** ${p.creature || '未设置'}
- **氛围:** ${p.vibes.length ? p.vibes.join(', ') : '未设置'}
- **Avatar:** ${p.avatar || 'brain'}
`
  }
  if (section === 'USER' || section === 'all') {
    const p = persona.user
    files.value.USER = `# 服务对象

- **名字:** ${p.name || '未设定'}
- **如何称呼:** ${p.callName || '未设定'}

## 背景

${p.context || '暂无背景信息。'}
`
  }
  if (section === 'SOUL' || section === 'all') {
    const p = persona.soul
    const formalDesc = p.formality > 70 ? '专业、严谨、权威' : p.formality > 30 ? '自然、灵活、平衡' : '随意、轻松、友善'
    const verbDesc = p.verbosity > 70 ? '详细当它重要时' : p.verbosity > 30 ? '简洁与详细平衡' : '极度精简，直击要点'
    const initDesc = p.initiative > 70 ? '主动探索和建议' : p.initiative > 30 ? '适度主动' : '等待指令再行动'
    const empDesc = p.empathy > 70 ? '富有共情，理解情感' : p.empathy > 30 ? '理性与共情并重' : '纯理性，就事论事'
    files.value.SOUL = `# 性格灵魂

## 风格调性

- **专业度 (${p.formality}/100):** ${formalDesc}
- **详细度 (${p.verbosity}/100):** ${verbDesc}
- **主动性 (${p.initiative}/100):** ${initDesc}
- **共情度 (${p.empathy}/100):** ${empDesc}

## 核心原则

${p.principles.length ? p.principles.map(x => '- ' + x).join('\n') : '- 未设定'}
`
  }
  if (section === 'TOOLS' || section === 'all') {
    const t = persona.tools
    const tendencyMap = { aggressive: '激进 — 直接执行，不问', balanced: '平衡 — 视情况而定', conservative: '保守 — 先问再执行' }
    let md = `# 工具箱

## 使用倾向

- **策略:** ${tendencyMap[t.tendency] || '平衡'}
`
    if (t.constraints.browser) md += `\n## 浏览器约束\n\n${t.constraints.browser}\n`
    if (t.constraints.exec) md += `\n## 终端约束\n\n${t.constraints.exec}\n`
    if (t.constraints.write) md += `\n## 写文件约束\n\n${t.constraints.write}\n`
    files.value.TOOLS = md
  }
  if (section === 'BOOTSTRAP' || section === 'all') {
    const b = persona.bootstrap
    let tasks = []
    if (b.greeting) tasks.push('主动向用户发送一条问候消息')
    if (b.scanProject) tasks.push('读取工作区下的所有项目目录结构')
    files.value.BOOTSTRAP = `# 入职自检

Agent 启动时自动执行的检查清单。

## 初始化动作

${tasks.map(t => '- [x] ' + t).join('\n') || '- [ ] 无'}

## 首次唤醒任务

${b.firstTask || '暂无自定义启动指令。'}
`
  }
  if (section === 'HEARTBEAT' || section === 'all') {
    const h = persona.heartbeat
    const goalMap = { summarize: '整理过去的对话摘要', todo: '检查是否有遗漏的待办事项', crawl: '主动上网爬取特定资讯', standby: '安静待机，不做任何动作' }
    const distDesc = h.disturbance > 70 ? '高“一有风吹草动立刻发消息”' : h.disturbance > 30 ? '中“重要事项才会打扰”' : '低“没重要的事也不发消息”'
    files.value.HEARTBEAT = `# 赛博心跳

定义 Agent 的周期性自主行为。

## 被唤醒时的目标

- **主要动作:** ${goalMap[h.wakeGoal] || '安静待机'}

## 主动打扰阈值

- **等级 (${h.disturbance}/100):** ${distDesc}
`
  }
  if (section === 'CONVENTIONS' || section === 'all') {
    const c = persona.conventions
    let md = '# 代码规范\n\n'
    if (c.languages.length) md += `## 语言与框架\n\n${c.languages.map(l => '- ' + l).join('\n')}\n\n`
    if (c.codeStyle) md += `## 代码风格\n\n- 遵循 **${c.codeStyle}** 规范\n\n`
    md += '## 排版要求\n\n'
    if (c.useMarkdown) md += '- 回复必须使用 Markdown 格式排版\n'
    if (c.showThinking) md += '- 在给出结论前展示思考过程\n'
    if (c.useMermaid) md += '- 解释复杂逻辑时附带 Mermaid 流程图\n'
    files.value.CONVENTIONS = md
  }
}

function parseMarkdownToPersona() {
  initializing.value = true
  // Parse IDENTITY
  const id = files.value.IDENTITY || ''
  const nameMatch = id.match(/\*\*名称.*?:\*\*\s*(.+)/)
  const creatureMatch = id.match(/\*\*类型.*?:\*\*\s*(.+)/)
  const vibeMatch = id.match(/\*\*氛围.*?:\*\*\s*(.+)/)
  const avatarMatch = id.match(/\*\*.*?Avatar.*?:\*\*\s*(.+)/)
  persona.identity.name = nameMatch?.[1]?.trim() || ''
  persona.identity.creature = creatureMatch?.[1]?.trim() || ''
  persona.identity.vibes = vibeMatch?.[1]?.split(/[,、]/).map(s => s.trim()).filter(Boolean) || []
  persona.identity.avatar = avatarMatch?.[1]?.trim() || 'brain'

  // Parse USER
  const usr = files.value.USER || ''
  const uNameMatch = usr.match(/\*\*名字.*?:\*\*\s*(.+)/)
  const uCallMatch = usr.match(/\*\*如何称呼.*?:\*\*\s*(.+)/)
  const uTzMatch = usr.match(/\*\*时区.*?:\*\*\s*(.+)/)
  persona.user.name = uNameMatch?.[1]?.trim() || ''
  persona.user.callName = uCallMatch?.[1]?.trim() || ''
  const bgMatch = usr.match(/## 背景\n\n([\s\S]*?)\n---/)
  persona.user.context = bgMatch?.[1]?.trim() || ''

  // Parse SOUL
  const soul = files.value.SOUL || ''
  const fmMatch = soul.match(/专业度\s*\((\d+)/)
  const vbMatch = soul.match(/详细度\s*\((\d+)/)
  const inMatch = soul.match(/主动性\s*\((\d+)/)
  const emMatch = soul.match(/共情度\s*\((\d+)/)
  persona.soul.formality = fmMatch ? parseInt(fmMatch[1]) : 50
  persona.soul.verbosity = vbMatch ? parseInt(vbMatch[1]) : 50
  persona.soul.initiative = inMatch ? parseInt(inMatch[1]) : 50
  persona.soul.empathy = emMatch ? parseInt(emMatch[1]) : 50
  const principlesSection = soul.match(/## 核心原则\n\n([\s\S]*?)\n---/)
  if (principlesSection) {
    persona.soul.principles = principlesSection[1].split('\n').map(l => l.replace(/^-\s*/, '').trim()).filter(Boolean)
  }
  // Allow watchers to fire again after next tick
  setTimeout(() => { initializing.value = false }, 0)
}

const anyModified = computed(() => tabs.some(t => isModified(t.key)))

async function saveAllVisual() {
  saving.value = true
  try {
    syncPersonaToMarkdown('all')
    const isMain = editingAgent.value && (editingAgent.value.role === 'main' || editingAgent.value.id === 'main')
    let sections = ['IDENTITY', 'USER', 'SOUL', 'TOOLS', 'BOOTSTRAP', 'HEARTBEAT', 'CONVENTIONS']
    if (!isMain) sections = sections.filter(s => s !== 'AGENTS')
    for (const name of sections) {
      if (isModified(name)) {
        await saveAgentFile({ agentId: editingAgent.value.id, name, content: files.value[name] })
        originals.value[name] = files.value[name]
      }
    }
    gm.success('人格已保存')
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

// Watch persona changes and sync to markdown
watch(() => persona.identity, () => syncPersonaToMarkdown('IDENTITY'), { deep: true })
watch(() => persona.user, () => syncPersonaToMarkdown('USER'), { deep: true })
watch(() => [persona.soul.formality, persona.soul.verbosity, persona.soul.initiative, persona.soul.empathy],
  () => syncPersonaToMarkdown('SOUL')
)
watch(() => persona.tools, () => syncPersonaToMarkdown('TOOLS'), { deep: true })
watch(() => persona.bootstrap, () => syncPersonaToMarkdown('BOOTSTRAP'), { deep: true })
watch(() => persona.heartbeat.disturbance, () => syncPersonaToMarkdown('HEARTBEAT'))
watch(() => persona.conventions, () => syncPersonaToMarkdown('CONVENTIONS'), { deep: true })

// ========== Agent 列表 ==========

async function fetchAgents() {
  loading.value = true
  try {
    const [agentsRes, tplRes, modelsRes] = await Promise.all([
      listAgents(), getAgentTemplates(), getConfiguredModels()
    ])
    if (agentsRes?.agents) {
      agents.value = agentsRes.agents
    }
    if (tplRes?.templates) {
      templates.value = tplRes.templates
    }
    if (modelsRes?.models) {
      modelOptions.value = modelsRes.models
    }
    if (modelsRes?.defaultModel) {
      defaultModel.value = modelsRes.defaultModel
    }
  } catch (e) {
    gm.error('加载失败: ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

function statusLabel(status) {
  const map = { idle: '待命', thinking: '思考中', acting: '执行中', error: '异常' }
  return map[status] || status
}

// ========== 创建/编辑 Modal ==========

const showModal = ref(false)
const formRef = ref(null)
const formData = ref({ id: '', name: '', avatar: 'brain', model: '', role: 'specialist', parentId: '', description: '' })

const modalTitle = computed(() => formData.value.id ? '编辑 Agent' : '新建 Agent')

const roleOptions = [
  { label: '主控 (Main)', value: 'main' },
  { label: '专员 (Specialist)', value: 'specialist' },
]

const parentOptions = computed(() =>
  agents.value
    .filter(a => a.role === 'main' && a.id !== formData.value.id)
    .map(a => ({ label: a.name, value: a.id }))
)

function openCreateModal() {
  formData.value = { id: '', name: '', avatar: 'brain', model: defaultModel.value, role: 'specialist', parentId: '', description: '' }
  // 默认从属于主 Agent
  const mainAgent = agents.value.find(a => a.role === 'main')
  if (mainAgent) formData.value.parentId = mainAgent.id
  showModal.value = true
}

function openEditModal(agent) {
  formData.value = { ...agent }
  showModal.value = true
}

async function submitForm() {
  if (!formData.value.name.trim()) {
    gm.error('请输入 Agent 名称')
    return
  }
  if (!/^[\u4e00-\u9fa5a-zA-Z0-9_-]+$/.test(formData.value.name)) {
    gm.error('Agent 名称仅支持中英文、数字、下划线和短横线')
    return
  }
  submitting.value = true
  try {
    if (formData.value.id) {
      await updateAgent(formData.value)
      gm.success('Agent 已更新')
    } else {
      await createAgent(formData.value)
      gm.success('Agent 已创建')
    }
    showModal.value = false
    await fetchAgents()
  } catch (e) {
    gm.error('操作失败: ' + (e.message || ''))
  } finally {
    submitting.value = false
  }
}

async function confirmDelete(agent) {
  const gmApi = gm.getGmApi()
  const doDelete = async () => {
    deleting.value = agent.id
    try {
      await deleteAgent({ id: agent.id })
      gm.success(`${agent.name} 已删除`)
      await fetchAgents()
    } catch (e) {
      gm.error('删除失败: ' + (e.message || ''))
    } finally {
      deleting.value = null
    }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '删除 Agent',
      content: `确定要删除「${agent.name}」吗？其人格文件将被永久清除。`,
      positiveText: '确定删除',
      negativeText: '取消',
      onPositiveClick: doDelete,
    })
  } else {
    if (confirm(`确定删除「${agent.name}」？`)) doDelete()
  }
}

// ========== Agent 详情编辑 ==========

const editingAgent = ref(null)
const activeTab = ref('IDENTITY')
const files = ref({ IDENTITY: '', USER: '', SOUL: '', AGENTS: '', TOOLS: '', BOOTSTRAP: '', HEARTBEAT: '', CONVENTIONS: '' })
const originals = ref({ IDENTITY: '', USER: '', SOUL: '', AGENTS: '', TOOLS: '', BOOTSTRAP: '', HEARTBEAT: '', CONVENTIONS: '' })
const editorRef = ref(null)

const tabs = [
  { key: 'IDENTITY', label: 'Identity', desc: '我是谁', icon: icons.fingerprint(16, 16) },
  { key: 'USER', label: 'User', desc: '你是谁', icon: icons.users(16, 16) },
  { key: 'SOUL', label: 'Soul', desc: '怎么聊', icon: icons.sparkles(16, 16) },
  { key: 'AGENTS', label: 'Agents', desc: '团队', icon: icons.robot(16, 16) },
  { key: 'TOOLS', label: 'Tools', desc: '工具箱', icon: icons.tool(16, 16) },
  { key: 'BOOTSTRAP', label: 'Bootstrap', desc: '入职', icon: icons.rocketSimple(16, 16) },
  { key: 'HEARTBEAT', label: 'Heartbeat', desc: '心跳', icon: icons.activity(16, 16) },
  { key: 'CONVENTIONS', label: 'Conventions', desc: '规范', icon: icons.fileCode(16, 16) },
  { key: 'MEMORY', label: 'Memory', desc: '记忆 · 可选', icon: icons.database(16, 16) },
]

const activeFileName = computed(() => {
  const found = filteredTabs.value.find(t => t.key === activeTab.value)
  return found ? found.key + '.md' : ''
})

// 根据当前编辑的 Agent 角色动态过滤 tabs
const filteredTabs = computed(() => {
  const isMain = editingAgent.value && (editingAgent.value.role === 'main' || editingAgent.value.id === 'main')
  if (isMain) return tabs
  return tabs.filter(t => t.key !== 'AGENTS')
})

const currentContent = computed(() => files.value[activeTab.value] || '')

function isModified(key) { return files.value[key] !== originals.value[key] }
function switchTab(key) { activeTab.value = key }
function onInput(e) { files.value[activeTab.value] = e.target.value }

async function openDetail(agent) {
  editingAgent.value = agent
  detailLoading.value = true
  activeTab.value = 'IDENTITY'
  editMode.value = 'visual'
  try {
    const res = await getAgentDetail({ id: agent.id })
    if (res?.files) {
      res.files.forEach(f => {
        files.value[f.name] = f.content
        originals.value[f.name] = f.content
      })
      parseMarkdownToPersona()
      // 用 agent 自身数据覆盖表单（名称和头像以 agent 记录为准）
      if (agent.name) persona.identity.name = agent.name
      if (agent.avatar) persona.identity.avatar = agent.avatar
    }
  } catch (e) {
    gm.error('加载详情失败: ' + (e.message || ''))
  } finally {
    detailLoading.value = false
  }
}

function exitDetail() {
  editingAgent.value = null
  const empty = { IDENTITY: '', USER: '', SOUL: '', AGENTS: '', TOOLS: '', BOOTSTRAP: '', HEARTBEAT: '', CONVENTIONS: '', MEMORY: '' }
  files.value = { ...empty }
  originals.value = { ...empty }
}

async function saveFile() {
  saving.value = true
  try {
    // 主 Agent 使用原有接口，其他 Agent 也使用原有接口（后端按 agentId 路由暂未实装，先用通用）
    if (editingAgent.value.id === 'main') {
      await saveAgentFile({
        agentId: 'main',
        name: activeTab.value,
        content: files.value[activeTab.value],
      })
    } else {
      await saveAgentFile({
        agentId: editingAgent.value.id,
        name: activeTab.value,
        content: files.value[activeTab.value],
      })
    }
    originals.value[activeTab.value] = files.value[activeTab.value]
    gm.success(`${activeFileName.value} 已保存`)
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function resetFile() {
  const gmApi = gm.getGmApi()
  const doReset = async () => {
    saving.value = true
    try {
      await resetAgentFile({ agentId: editingAgent.value.id, name: activeTab.value })
      const res = await getAgentDetail({ id: editingAgent.value.id })
      if (res?.files) {
        res.files.forEach(f => {
          files.value[f.name] = f.content
          originals.value[f.name] = f.content
        })
      }
      gm.success(`${activeFileName.value} 已恢复默认`)
    } catch (e) {
      gm.error('重置失败: ' + (e.message || ''))
    } finally {
      saving.value = false
    }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '恢复默认',
      content: `确定要将 ${activeFileName.value} 恢复为默认内容吗？当前内容将被覆盖。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: doReset,
    })
  } else {
    if (confirm(`确定恢复 ${activeFileName.value} 为默认内容？`)) doReset()
  }
}

// ========== 模板 ==========

const templateDropdownOptions = computed(() =>
  templates.value.map(t => ({ label: t.name, key: t.key }))
)

async function onTemplateSelect(key) {
  const tpl = templates.value.find(t => t.key === key)
  if (!tpl) return
  const doApply = async () => {
    saving.value = true
    try {
      await applyAgentTemplate({ key })
      const res = await getAgentDetail({ id: editingAgent.value.id })
      if (res?.files) {
        res.files.forEach(f => {
          files.value[f.name] = f.content
          originals.value[f.name] = f.content
        })
      }
      gm.success(`已应用「${tpl.name}」模板`)
    } catch (e) {
      gm.error('应用模板失败: ' + (e.message || ''))
    } finally {
      saving.value = false
    }
  }
  const gmApi = gm.getGmApi()
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '应用模板',
      content: `确定应用「${tpl.name}」模板吗？这将覆盖 IDENTITY.md、USER.md、SOUL.md 三个文件的内容。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: doApply,
    })
  } else {
    if (confirm(`确定应用「${tpl.name}」模板？`)) doApply()
  }
}

onMounted(() => {
  fetchAgents()
})
</script>

<style scoped>
.agents-page {
  width: 100%; height: 100%;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
}
.agents-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  min-height: 0;
  position: relative;
}

/* ===== Header ===== */
.agents-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--jm-accent-7);
  margin: 0;
}
.header-hint {
  font-size: 12px;
  color: var(--jm-accent-4);
  padding-left: 28px;
}
.create-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: 8px;
  border: 1px solid var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
  color: var(--jm-primary-2);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.create-btn:hover {
  background: rgba(var(--jm-primary-1-rgb), 0.18);
  box-shadow: 0 0 12px rgba(var(--jm-primary-1-rgb), 0.2);
}
.refresh-btn {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); cursor: pointer; transition: all 0.2s;
}
.refresh-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }
.refresh-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.spinning { animation: spin 0.8s linear infinite; }

/* ===== Cyber Grid Background ===== */
.cyber-grid-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background-image:
    linear-gradient(rgba(var(--jm-primary-1-rgb), 0.015) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--jm-primary-1-rgb), 0.015) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse 70% 60% at 50% 30%, black 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(ellipse 70% 60% at 50% 30%, black 40%, transparent 100%);
}

/* ===== List Toolbar ===== */
.list-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  position: relative;
  z-index: 1;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  border-radius: 10px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  backdrop-filter: blur(8px);
  flex: 1;
  max-width: 320px;
  transition: all 0.25s;
}
.search-box:focus-within {
  border-color: var(--jm-primary-2);
  box-shadow: 0 0 16px rgba(var(--jm-primary-1-rgb), 0.12), 0 4px 12px rgba(0,0,0,0.06);
  background: rgba(var(--jm-accent-1-rgb), 0.6);
}
.search-icon { display: flex; color: var(--jm-accent-4); flex-shrink: 0; }
.search-input {
  border: none; outline: none; background: transparent;
  color: var(--jm-accent-7); font-size: 12px; width: 100%;
}
.search-input::placeholder { color: var(--jm-accent-3); }
.toolbar-actions { display: flex; gap: 8px; }

/* ===== List Layout (Grid + Side Panel) ===== */
.list-layout {
  display: grid;
  grid-template-columns: 1fr 260px;
  gap: 16px;
  flex: 1;
  min-height: 0;
  position: relative;
  z-index: 1;
}

/* ===== Card Grid ===== */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
  align-content: start;
}

/* ===== Agent Bento Card ===== */
.agent-card {
  position: relative;
  border-radius: 14px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}
.agent-card:hover {
  border-color: rgba(var(--jm-primary-1-rgb), 0.4);
  background: rgba(var(--jm-accent-1-rgb), 0.5);
  box-shadow:
    0 8px 32px rgba(0,0,0,0.08),
    0 0 20px rgba(var(--jm-primary-1-rgb), 0.06);
  transform: translateY(-2px);
}
.agent-card.is-main {
  border-color: rgba(var(--jm-primary-1-rgb), 0.25);
  background: rgba(var(--jm-primary-1-rgb), 0.04);
}
.agent-card.is-main:hover {
  border-color: rgba(var(--jm-primary-1-rgb), 0.5);
  box-shadow:
    0 8px 32px rgba(0,0,0,0.1),
    0 0 24px rgba(var(--jm-primary-1-rgb), 0.1);
}

/* Card glow bar */
.card-glow-bar {
  position: absolute;
  top: 0; left: 16px; right: 16px;
  height: 2px;
  border-radius: 0 0 2px 2px;
  background: var(--jm-accent-3);
  opacity: 0.4;
  transition: all 0.3s;
}
.card-glow-bar.idle { background: #4ade80; opacity: 0.6; box-shadow: 0 0 8px rgba(74,222,128,0.3); }
.card-glow-bar.thinking { background: #60a5fa; opacity: 0.8; box-shadow: 0 0 12px rgba(96,165,250,0.4); animation: bar-pulse 1.5s ease-in-out infinite; }
.card-glow-bar.acting { background: #fb923c; opacity: 0.8; box-shadow: 0 0 12px rgba(251,146,60,0.4); animation: bar-pulse 0.8s ease-in-out infinite; }
.card-glow-bar.error { background: #f87171; opacity: 0.9; box-shadow: 0 0 12px rgba(248,113,113,0.5); }
@keyframes bar-pulse { 0%,100% { opacity: 0.5; } 50% { opacity: 1; } }

/* Card header */
.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

/* Avatar with halo */
.card-avatar {
  position: relative;
  width: 40px; height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  border: 1px solid var(--jm-glass-border);
  flex-shrink: 0;
  transition: all 0.3s;
}
.card-avatar .avatar-icon {
  display: flex; color: var(--jm-accent-5);
  position: relative; z-index: 1;
  transition: color 0.3s;
}
.agent-card:hover .card-avatar .avatar-icon { color: var(--jm-primary-1); }
.agent-card.is-main .card-avatar .avatar-icon { color: var(--jm-primary-1); }
.agent-card.is-main .card-avatar {
  border-color: var(--jm-glass-border-hover);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
}

/* Avatar halo (glow ring for active states) */
.avatar-halo {
  position: absolute; inset: -4px;
  border-radius: 14px;
  border: 1.5px solid transparent;
  opacity: 0;
  transition: all 0.3s;
}
.avatar-halo.idle {
  border-color: rgba(74,222,128,0.3);
  box-shadow: 0 0 10px rgba(74,222,128,0.15);
  opacity: 1;
}
.avatar-halo.thinking {
  border-color: rgba(96,165,250,0.5);
  box-shadow: 0 0 14px rgba(96,165,250,0.25);
  opacity: 1;
  animation: halo-breathe 2s ease-in-out infinite;
}
.avatar-halo.acting {
  border-color: rgba(251,146,60,0.5);
  box-shadow: 0 0 14px rgba(251,146,60,0.25);
  opacity: 1;
  animation: halo-breathe 1s ease-in-out infinite;
}
.avatar-halo.error {
  border-color: rgba(248,113,113,0.6);
  box-shadow: 0 0 14px rgba(248,113,113,0.3);
  opacity: 1;
}
@keyframes halo-breathe {
  0%,100% { opacity: 0.6; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.08); }
}

/* Card title area */
.card-title-area { flex: 1; min-width: 0; }
.card-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--jm-accent-7);
}
.card-model {
  font-size: 10px;
  color: var(--jm-accent-4);
  margin-top: 2px;
  font-family: 'SF Mono','Fira Code',monospace;
}

/* Card actions */
.card-actions {
  display: flex; gap: 4px;
  opacity: 0; transition: opacity 0.2s;
  flex-shrink: 0;
}
.agent-card:hover .card-actions { opacity: 1; }

/* Card description */
.card-desc {
  font-size: 11px;
  color: var(--jm-accent-4);
  line-height: 1.5;
  margin-bottom: 10px;
  display: -webkit-box;
  line-clamp: 2;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Status badge */
.card-status-row { margin-bottom: 10px; }
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.5px;
}
.status-dot-mini {
  width: 5px; height: 5px;
  border-radius: 50%;
  background: var(--jm-accent-3);
}
.status-badge.idle {
  background: rgba(74,222,128,0.1);
  color: #4ade80;
}
.status-badge.idle .status-dot-mini { background: #4ade80; box-shadow: 0 0 4px #4ade80; }
.status-badge.thinking {
  background: rgba(96,165,250,0.1);
  color: #60a5fa;
}
.status-badge.thinking .status-dot-mini { background: #60a5fa; box-shadow: 0 0 4px #60a5fa; animation: pulse-glow 1.5s ease-in-out infinite; }
.status-badge.acting {
  background: rgba(251,146,60,0.1);
  color: #fb923c;
}
.status-badge.acting .status-dot-mini { background: #fb923c; box-shadow: 0 0 4px #fb923c; animation: pulse-glow 0.8s ease-in-out infinite; }
.status-badge.error {
  background: rgba(248,113,113,0.1);
  color: #f87171;
}
.status-badge.error .status-dot-mini { background: #f87171; box-shadow: 0 0 6px #f87171; }

@keyframes pulse-glow {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* Card stats */
.card-stats {
  display: flex;
  gap: 2px;
  margin-bottom: 10px;
}
.stat-mini {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 0;
  border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
}
.stat-mini-val {
  font-size: 14px;
  font-weight: 700;
  color: var(--jm-accent-7);
  font-family: 'SF Mono','Fira Code',monospace;
}
.stat-mini-label {
  font-size: 9px;
  color: var(--jm-accent-4);
  margin-top: 1px;
}

/* Capability tags */
.card-capabilities {
  display: flex;
  gap: 4px;
}
.cap-tag {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px; height: 26px;
  border-radius: 6px;
  background: rgba(var(--jm-accent-1-rgb), 0.5);
  border: 1px solid var(--jm-glass-border);
  color: var(--jm-accent-4);
  transition: all 0.2s;
}
.cap-tag:hover {
  color: var(--jm-primary-1);
  border-color: var(--jm-glass-border-hover);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  box-shadow: 0 0 8px rgba(var(--jm-primary-1-rgb), 0.15);
}

/* Empty search */
.empty-search {
  grid-column: 1 / -1;
  text-align: center;
  padding: 40px 20px;
  color: var(--jm-accent-4);
  font-size: 13px;
}

/* ===== Team Panel ===== */
.team-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: sticky;
  top: 0;
  height: fit-content;
}
.panel-section {
  border-radius: 14px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  padding: 16px;
}
.panel-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--jm-accent-5);
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 12px;
}

/* Ring chart */
.ring-chart-wrap {
  position: relative;
  width: 100px; height: 100px;
  margin: 0 auto 8px;
}
.ring-chart { width: 100%; height: 100%; }
.ring-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}
.ring-val {
  font-size: 22px;
  font-weight: 700;
  color: var(--jm-primary-1);
  font-family: 'SF Mono','Fira Code',monospace;
}
.ring-sub {
  font-size: 12px;
  color: var(--jm-accent-4);
  font-family: 'SF Mono','Fira Code',monospace;
}
.ring-hint {
  text-align: center;
  font-size: 11px;
  color: var(--jm-accent-4);
}

/* Team composition */
.team-comp {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.comp-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.comp-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.comp-dot.main { background: var(--jm-primary-1); box-shadow: 0 0 4px var(--jm-primary-1); }
.comp-dot.specialist { background: #a78bfa; box-shadow: 0 0 4px #a78bfa; }
.comp-dot.idle { background: #4ade80; box-shadow: 0 0 4px #4ade80; }
.comp-dot.thinking { background: #60a5fa; box-shadow: 0 0 4px #60a5fa; }
.comp-dot.acting { background: #fb923c; box-shadow: 0 0 4px #fb923c; }
.comp-dot.error { background: #f87171; box-shadow: 0 0 4px #f87171; }
.comp-label { flex: 1; color: var(--jm-accent-5); }
.comp-val {
  font-weight: 700;
  color: var(--jm-accent-7);
  font-family: 'SF Mono','Fira Code',monospace;
  min-width: 20px;
  text-align: right;
}

/* Activity feed */
.feed-section { flex: 1; min-height: 0; }
.activity-feed {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}
.feed-item {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 11px;
  line-height: 1.4;
}
.feed-time {
  color: var(--jm-accent-4);
  font-family: 'SF Mono','Fira Code',monospace;
  font-size: 10px;
  flex-shrink: 0;
  min-width: 36px;
}
.feed-dot {
  width: 5px; height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 4px;
  background: var(--jm-accent-3);
}
.feed-dot.idle { background: #4ade80; }
.feed-dot.thinking { background: #60a5fa; }
.feed-dot.acting { background: #fb923c; }
.feed-dot.error { background: #f87171; }
.feed-text { color: var(--jm-accent-5); }

/* Role Tag (inline) */
.role-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 500;
  flex-shrink: 0;
}
.role-tag.main {
  background: rgba(var(--jm-primary-1-rgb), 0.12);
  color: var(--jm-primary-1);
}
.role-tag.specialist {
  background: rgba(var(--jm-accent-1-rgb), 0.5);
  color: var(--jm-accent-5);
}
.model-tag {
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 9px;
  font-weight: 400;
  color: var(--jm-accent-4);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  margin-left: 4px;
  flex-shrink: 0;
}

/* Row Actions */
.row-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}
.agent-row:hover .row-actions { opacity: 1; }

.act-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px; height: 26px;
  border-radius: 5px;
  border: 1px solid var(--jm-glass-border);
  background: transparent;
  color: var(--jm-accent-5);
  cursor: pointer;
  transition: all 0.15s;
}
.act-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }
.act-btn.danger:hover { border-color: #f87171; color: #f87171; }

/* ===== Detail View ===== */
.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border-radius: 6px;
  border: 1px solid var(--jm-glass-border);
  background: transparent;
  color: var(--jm-accent-5);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.back-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }

.detail-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}
.detail-icon { display: flex; align-items: center; color: var(--jm-accent-5); }
.detail-name { font-size: 16px; font-weight: 600; color: var(--jm-accent-7); }

/* ===== Persona Recommend Cards ===== */
.persona-recommend {
  margin-bottom: 8px;
}
.recommend-title {
  font-size: 11px;
  color: var(--jm-accent-4);
  margin-bottom: 6px;
  letter-spacing: 1px;
}
.recommend-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.recommend-card {
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  cursor: pointer;
  transition: all 0.25s;
  position: relative;
  overflow: hidden;
}
.recommend-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  opacity: 0;
  transition: opacity 0.3s;
}
.recommend-card:hover { border-color: var(--jm-accent-3); }
.recommend-card:hover::before { opacity: 1; }
.recommend-card.active::before { opacity: 1; }

/* Theme: Red (铁血教官) */
.recommend-card.theme-red::before { background: #f87171; }
.recommend-card.theme-red:hover { border-color: rgba(248, 113, 113, 0.4); box-shadow: 0 0 12px rgba(248, 113, 113, 0.08); }
.recommend-card.theme-red.active { border-color: rgba(248, 113, 113, 0.5); background: rgba(248, 113, 113, 0.04); }
.recommend-card.theme-red .rc-icon { color: #f87171; }

/* Theme: Purple (灰度博弈者) */
.recommend-card.theme-purple::before { background: #a78bfa; }
.recommend-card.theme-purple:hover { border-color: rgba(167, 139, 250, 0.4); box-shadow: 0 0 12px rgba(167, 139, 250, 0.08); }
.recommend-card.theme-purple.active { border-color: rgba(167, 139, 250, 0.5); background: rgba(167, 139, 250, 0.04); }
.recommend-card.theme-purple .rc-icon { color: #a78bfa; }

/* Theme: Neon (混沌创意官) */
.recommend-card.theme-neon::before { background: linear-gradient(90deg, #22d3ee, #a78bfa, #f472b6); }
.recommend-card.theme-neon:hover { border-color: rgba(34, 211, 238, 0.4); box-shadow: 0 0 12px rgba(34, 211, 238, 0.08); }
.recommend-card.theme-neon.active { border-color: rgba(34, 211, 238, 0.5); background: rgba(34, 211, 238, 0.03); }
.recommend-card.theme-neon .rc-icon { color: #22d3ee; }

.rc-icon {
  display: flex;
  margin-bottom: 6px;
  color: var(--jm-accent-4);
}
.rc-info { margin-bottom: 4px; }
.rc-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--jm-accent-7);
}
.rc-sub {
  font-size: 9px;
}

/* ========== Avatar Grid ========== */
.avatar-grid {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: 8px;
  width: 100%;
}
.avatar-option {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  border-radius: 8px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  color: var(--jm-accent-5);
  cursor: pointer;
  transition: all 0.2s;
}
.avatar-option:hover {
  border-color: var(--jm-accent-3);
  color: var(--jm-accent-6);
  background: rgba(var(--jm-accent-2-rgb), 0.5);
}
.avatar-option.active {
  border-color: var(--jm-primary-1);
  color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  box-shadow: 0 0 12px rgba(var(--jm-primary-1-rgb), 0.2);
}

.modal-avatar-grid {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: 6px;
  width: 100%;
}
.modal-avatar-option {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  border-radius: 6px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  color: var(--jm-accent-5);
  cursor: pointer;
  transition: all 0.2s;
}
.modal-avatar-option:hover {
  border-color: var(--jm-accent-3);
  color: var(--jm-accent-6);
  background: rgba(var(--jm-accent-2-rgb), 0.5);
}
.modal-avatar-option.active {
  border-color: var(--jm-primary-1);
  color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  box-shadow: 0 0 12px rgba(var(--jm-primary-1-rgb), 0.2);
}
.rc-desc {
  font-size: 10px;
  color: var(--jm-accent-5);
  line-height: 1.4;
}

/* ===== Soul Inject Title ===== */
.soul-inject-title {
  position: relative;
  font-size: 13px;
  font-weight: 600;
  color: var(--jm-accent-5);
  letter-spacing: 2px;
  padding: 10px 0 6px 18px;
}
.soul-glow {
  position: absolute;
  left: 0; top: 50%;
  width: 4px; height: 16px;
  border-radius: 2px;
  background: var(--jm-primary-1);
  transform: translateY(-50%);
  box-shadow: 0 0 8px var(--jm-primary-1);
  animation: glow-pulse 2s ease-in-out infinite;
}
@keyframes glow-pulse {
  0%, 100% { opacity: 0.6; box-shadow: 0 0 4px var(--jm-primary-1); }
  50% { opacity: 1; box-shadow: 0 0 12px var(--jm-primary-1); }
}

/* ===== Mode Toggle ===== */
.mode-toggle {
  display: flex;
  gap: 2px;
  padding: 2px;
  border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 1px solid var(--jm-glass-border);
  width: fit-content;
  margin-bottom: 10px;
}
.mode-toggle button {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--jm-accent-4);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.mode-toggle button.active {
  background: rgba(var(--jm-primary-1-rgb), 0.12);
  color: var(--jm-primary-1);
  box-shadow: 0 0 6px rgba(var(--jm-primary-1-rgb), 0.15);
}
.mode-toggle button:hover:not(.active) { color: var(--jm-accent-6); }

/* ===== Persona Sections ===== */
.persona-section {
  border-radius: 10px;
  border: 1px solid var(--jm-glass-border);
  background: var(--jm-glass-bg);
  margin-bottom: 8px;
  transition: all 0.25s;
}
.persona-section:hover {
  border-color: var(--jm-accent-3);
}
.persona-section.active {
  border-color: rgba(var(--jm-primary-1-rgb), 0.4);
  background: rgba(var(--jm-primary-1-rgb), 0.03);
  box-shadow: 0 0 16px rgba(var(--jm-primary-1-rgb), 0.06), inset 0 0 20px rgba(var(--jm-primary-1-rgb), 0.02);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
}
.section-icon { display: flex; color: var(--jm-accent-4); }
.persona-section.active .section-icon { color: var(--jm-primary-1); }
.section-title { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); }
.section-hint { font-size: 11px; color: var(--jm-accent-4); margin-left: auto; }
.section-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: var(--jm-primary-1);
  box-shadow: 0 0 4px var(--jm-primary-1);
}
.section-header {
  cursor: pointer;
  user-select: none;
}
.section-chevron {
  display: flex;
  margin-left: 6px;
  transition: transform 0.25s ease;
  transform: rotate(-90deg);
  opacity: 0.4;
  color: var(--jm-accent-5);
}
.section-chevron.open {
  transform: rotate(0deg);
  opacity: 0.7;
}

.section-body {
  padding: 0 14px 14px;
}

/* ===== Check Group ===== */
.check-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.check-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--jm-accent-6);
  cursor: pointer;
}
.check-item input[type="checkbox"] {
  accent-color: #5569FA;
  width: 14px;
  height: 14px;
}

/* ===== Field Rows ===== */
.field-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 8px;
}
.field-row label {
  min-width: 56px;
  font-size: 12px;
  color: var(--jm-accent-5);
  line-height: 28px;
  flex-shrink: 0;
  text-align: right;
}

/* ===== Slider Rows ===== */
.slider-row {
  margin-bottom: 10px;
}
.slider-row label {
  display: block;
  font-size: 12px;
  color: var(--jm-accent-5);
  margin-bottom: 2px;
}
.slider-labels {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--jm-accent-4);
  margin-bottom: 4px;
}

/* ===== Vibe Tags ===== */
.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  flex: 1;
}
.vibe-tag {
  padding: 3px 10px;
  border-radius: 14px;
  border: 1px solid var(--jm-glass-border);
  background: transparent;
  color: var(--jm-accent-5);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}
.vibe-tag:hover { border-color: var(--jm-accent-4); color: var(--jm-accent-7); }
.vibe-tag.selected {
  border-color: var(--jm-primary-1);
  color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
}
/* Glow effects */
.vibe-tag.glow-blue.selected { border-color: #60a5fa; color: #60a5fa; background: rgba(96,165,250,0.08); box-shadow: 0 0 6px rgba(96,165,250,0.2); }
.vibe-tag.glow-orange.selected { border-color: #fb923c; color: #fb923c; background: rgba(251,146,60,0.08); box-shadow: 0 0 6px rgba(251,146,60,0.2); }
.vibe-tag.glow-green.selected { border-color: #4ade80; color: #4ade80; background: rgba(74,222,128,0.08); box-shadow: 0 0 6px rgba(74,222,128,0.2); }
.vibe-tag.glow-purple.selected { border-color: #a78bfa; color: #a78bfa; background: rgba(167,139,250,0.08); box-shadow: 0 0 6px rgba(167,139,250,0.2); }
.vibe-tag.glow-yellow.selected { border-color: #fbbf24; color: #fbbf24; background: rgba(251,191,36,0.08); box-shadow: 0 0 6px rgba(251,191,36,0.2); }
.vibe-tag.glow-cyan.selected { border-color: #22d3ee; color: #22d3ee; background: rgba(34,211,238,0.08); box-shadow: 0 0 6px rgba(34,211,238,0.2); }
.vibe-tag.glow-pink.selected { border-color: #f472b6; color: #f472b6; background: rgba(244,114,182,0.08); box-shadow: 0 0 6px rgba(244,114,182,0.2); }

/* ===== Save Bar ===== */
.save-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0 0;
  margin-top: 4px;
}
.save-bar-left { flex: 1; }
.save-bar-right { display: flex; gap: 6px; }

/* ===== Slider Description ===== */
.slider-desc {
  font-size: 10px;
  color: var(--jm-accent-4);
  margin-top: 2px;
  font-style: italic;
  transition: color 0.3s;
}
.slider-code {
  font-family: 'Courier New', monospace;
  font-size: 9px;
  color: #00d4ff;
  text-shadow: 0 0 6px rgba(0, 212, 255, 0.4);
  margin-left: 4px;
  font-weight: 400;
}

/* ===== Cyber Focus Effects ===== */
.editor-textarea:focus {
  border-color: rgba(0, 212, 255, 0.5);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.15), inset 0 0 12px rgba(0, 212, 255, 0.03);
}
.tab-item:hover .tab-icon {
  filter: drop-shadow(0 0 4px rgba(0, 212, 255, 0.5));
  transition: filter 0.2s;
}
.section-header:hover .section-icon {
  filter: drop-shadow(0 0 4px rgba(0, 212, 255, 0.5));
  transition: filter 0.2s;
}
.modified-hint {
  animation: cyber-pulse 2s ease-in-out infinite;
}
@keyframes cyber-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* ===== Preview Overlay ===== */
.preview-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 1000;
  background: var(--jm-overlay-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
}
.preview-panel {
  width: 560px;
  max-height: 70vh;
  border-radius: 12px;
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.3);
  background: rgba(var(--jm-accent-1-rgb), 0.85);
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 40px var(--jm-shadow-color), 0 0 20px rgba(var(--jm-primary-1-rgb), 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--jm-accent-7);
  border-bottom: 1px solid var(--jm-accent-2);
}
.preview-close {
  background: none; border: none;
  color: var(--jm-accent-4);
  cursor: pointer;
  font-size: 14px;
}
.preview-close:hover { color: var(--jm-accent-7); }
.preview-tabs {
  display: flex;
  gap: 2px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--jm-accent-2);
}
.preview-tabs button {
  padding: 4px 12px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: var(--jm-accent-4);
  font-size: 11px;
  cursor: pointer;
}
.preview-tabs button.active {
  background: rgba(var(--jm-primary-1-rgb), 0.12);
  color: var(--jm-primary-1);
}
.preview-content {
  padding: 16px;
  font-size: 12px;
  color: var(--jm-accent-6);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-y: auto;
  flex: 1;
  font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
}

/* ===== Tab Bar ===== */
.tab-bar {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  padding-bottom: 8px;
}
.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--jm-glass-border);
  background: transparent;
  color: var(--jm-accent-5);
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  flex-shrink: 0;
  min-width: 68px;
}
.tab-item:hover { border-color: var(--jm-accent-3); background: rgba(var(--jm-accent-1-rgb), 0.6); }
.tab-item.active { border-color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.08); color: var(--jm-primary-1); }
.tab-icon { display: flex; align-items: center; }
.tab-label { font-size: 13px; font-weight: 600; }
.tab-desc { font-size: 11px; opacity: 0.7; }
.tab-dot {
  position: absolute; top: 8px; right: 8px;
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--jm-primary-1);
}

/* ===== Editor ===== */
.editor-panel {
  border: 1px solid var(--jm-glass-border);
  border-radius: 10px;
  overflow: hidden;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  display: flex;
  flex-direction: column;
}
.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  border-bottom: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
}
.toolbar-left { display: flex; align-items: center; gap: 8px; }
.file-badge {
  font-size: 12px; font-weight: 600;
  color: var(--jm-accent-5);
  padding: 2px 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  border-radius: 4px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.modified-hint { font-size: 11px; color: var(--jm-primary-1); font-weight: 500; }
.toolbar-right { display: flex; align-items: center; gap: 6px; }

.tool-btn {
  display: flex; align-items: center; gap: 4px;
  padding: 5px 10px;
  border-radius: 6px;
  border: 1px solid var(--jm-glass-border);
  background: transparent;
  color: var(--jm-accent-5);
  font-size: 11px; cursor: pointer;
  transition: all 0.15s;
}
.tool-btn:hover:not(:disabled) { border-color: var(--jm-accent-3); color: var(--jm-accent-6); }
.tool-btn:disabled { opacity: 0.35; cursor: not-allowed; }

.save-btn {
  border-color: var(--jm-primary-2);
  color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
}
.save-btn:hover:not(:disabled) { background: rgba(var(--jm-primary-1-rgb), 0.15); }

.editor-wrapper { display: flex; flex: 1; }
.editor-textarea {
  width: 100%;
  min-height: 420px;
  padding: 16px;
  border: none; outline: none;
  resize: vertical;
  background: transparent;
  color: var(--jm-accent-7);
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.7;
  tab-size: 2;
}
.editor-textarea::placeholder { color: var(--jm-accent-3); }

/* ===== Template Button ===== */
.tpl-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-6);
  font-size: 12px; cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.tpl-btn:hover { border-color: var(--jm-primary-2); color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.08); }

/* ===== Loading ===== */
.loading-state, .editor-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  gap: 12px;
  color: var(--jm-accent-4);
  font-size: 13px;
}
.loading-spinner {
  width: 24px; height: 24px;
  border: 2px solid var(--jm-accent-2);
  border-top-color: var(--jm-primary-1);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* ===== Modal Footer ===== */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin 1s linear infinite; }

/* Dropdown overrides */
:deep(.n-dropdown-menu) {
  background: var(--jm-bg-color) !important;
  border: 1px solid var(--jm-glass-border) !important;
  border-radius: 8px !important;
}
:deep(.n-dropdown-option-body) { color: var(--jm-accent-6) !important; font-size: 13px !important; }
:deep(.n-dropdown-option-body:hover),
:deep(.n-dropdown-option-body--pending) {
  background: rgba(var(--jm-primary-1-rgb), 0.12) !important;
  color: var(--jm-primary-1) !important;
}

/* ===== Mode Switch ===== */
.mode-switch-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: 8px;
  border: 1px solid var(--jm-glass-border); background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-5); font-size: 12px; cursor: pointer; transition: all 0.2s; white-space: nowrap;
}
.mode-switch-btn:hover { border-color: var(--jm-primary-2); color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.06); }

/* ===== Cyber Topology ===== */
.topology-viewport {
  flex: 1; position: relative; border-radius: 12px;
  border: 1px solid var(--jm-glass-border); background: rgba(var(--jm-accent-1-rgb), 0.2);
  overflow: hidden;
}
.grid-bg {
  position: absolute; inset: 0;
  background-image: linear-gradient(rgba(var(--jm-primary-1-rgb),0.03) 1px,transparent 1px), linear-gradient(90deg,rgba(var(--jm-primary-1-rgb),0.03) 1px,transparent 1px);
  background-size: 40px 40px;
}
.scan-line {
  position: absolute; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg,transparent 0%,rgba(var(--jm-primary-1-rgb),0.1) 20%,rgba(var(--jm-primary-1-rgb),0.4) 50%,rgba(var(--jm-primary-1-rgb),0.1) 80%,transparent 100%);
  animation: scan 4s linear infinite; pointer-events: none;
}
@keyframes scan { 0% { top: -2px; } 100% { top: 100%; } }
.connections-layer { position: absolute; inset: 0; width: 100%; height: 100%; pointer-events: none; }
.agent-node {
  position: absolute; display: flex; flex-direction: column; align-items: center;
  gap: 4px; cursor: pointer; transition: transform 0.3s; z-index: 2;
}
.agent-node:hover { transform: scale(1.08); }
.pulse-ring {
  position: absolute; top: 50%; left: 50%; width: 70px; height: 70px;
  margin: -35px 0 0 -35px; border-radius: 50%; border: 1px solid var(--jm-primary-1);
  opacity: 0; animation: pulse-ring 3s ease-out infinite; pointer-events: none;
}
.pulse-ring.delay { animation-delay: 1.5s; }
.pulse-ring.idle { border-color: #4ade80; }
.pulse-ring.thinking { border-color: #60a5fa; animation-duration: 1.5s; }
.pulse-ring.acting { border-color: #fb923c; animation-duration: 1s; }
.pulse-ring.error { border-color: #f87171; animation-duration: 0.8s; }
@keyframes pulse-ring { 0% { transform: scale(0.8); opacity: 0.6; } 100% { transform: scale(2.2); opacity: 0; } }
.agent-node.main .pulse-ring { width: 90px; height: 90px; margin: -45px 0 0 -45px; }
.node-core {
  width: 64px; height: 64px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  border: 2px solid rgba(var(--jm-primary-1-rgb),0.4); background: rgba(var(--jm-accent-1-rgb),0.8);
  backdrop-filter: blur(8px); position: relative; z-index: 1; transition: all 0.3s;
}
.agent-node:hover .node-core { border-color: var(--jm-primary-1); box-shadow: 0 0 20px rgba(var(--jm-primary-1-rgb),0.3); }
.agent-node.main .node-core {
  width: 80px; height: 80px; border-color: var(--jm-primary-1);
  box-shadow: 0 0 20px rgba(var(--jm-primary-1-rgb),0.2), inset 0 0 15px rgba(var(--jm-primary-1-rgb),0.05);
}
.agent-node.idle .node-core { border-color: rgba(74,222,128,0.4); }
.agent-node.thinking .node-core { border-color: rgba(96,165,250,0.6); box-shadow: 0 0 15px rgba(96,165,250,0.2); animation: think-pulse 1.5s ease-in-out infinite; }
.agent-node.acting .node-core { border-color: rgba(251,146,60,0.6); box-shadow: 0 0 15px rgba(251,146,60,0.2); }
.agent-node.error .node-core { border-color: rgba(248,113,113,0.8); box-shadow: 0 0 15px rgba(248,113,113,0.3); animation: error-glitch 0.5s ease-in-out infinite; }
@keyframes think-pulse { 0%,100% { box-shadow: 0 0 15px rgba(96,165,250,0.2); } 50% { box-shadow: 0 0 30px rgba(96,165,250,0.4); } }
@keyframes error-glitch { 0%,100% { transform: none; } 25% { transform: translate(1px,-1px); } 50% { transform: translate(-1px,1px); } 75% { transform: translate(1px,1px); } }
.node-icon { display: flex; align-items: center; justify-content: center; color: var(--jm-accent-5); }
.agent-node.main .node-icon { color: var(--jm-primary-1); }
.agent-node:hover .node-icon { color: var(--jm-primary-1); }
.node-label { font-size: 12px; font-weight: 600; color: var(--jm-accent-7); text-shadow: 0 0 8px rgba(0,0,0,0.8); white-space: nowrap; }
.node-status-text { font-size: 9px; font-weight: 500; letter-spacing: 1px; text-transform: uppercase; color: var(--jm-accent-4); font-family: 'SF Mono','Fira Code',monospace; }
.agent-node.idle .node-status-text { color: #4ade80; }
.agent-node.thinking .node-status-text { color: #60a5fa; }
.agent-node.acting .node-status-text { color: #fb923c; }
.agent-node.error .node-status-text { color: #f87171; }
.empty-state { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; color: var(--jm-accent-4); font-size: 14px; }
.stats-bar {
  display: flex; align-items: center; justify-content: center; gap: 24px;
  padding: 10px 20px; border-radius: 10px; border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb),0.3); flex-shrink: 0;
}
.stat-item { display: flex; align-items: center; gap: 6px; }
.stat-value { font-size: 16px; font-weight: 700; color: var(--jm-accent-7); font-family: 'SF Mono','Fira Code',monospace; }
.stat-label { font-size: 11px; color: var(--jm-accent-4); }
.stat-dot { width: 8px; height: 8px; border-radius: 50%; }
.stat-dot.idle { background: #4ade80; box-shadow: 0 0 4px #4ade80; }
.stat-dot.thinking { background: #60a5fa; box-shadow: 0 0 4px #60a5fa; }
.stat-dot.acting { background: #fb923c; box-shadow: 0 0 4px #fb923c; }
.stat-dot.error { background: #f87171; box-shadow: 0 0 4px #f87171; }
</style>
