package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"guanxi/eazy-claw/internal/dto"
)

// AgentService Agent管理服务
type AgentService struct{}

// NewAgentService 创建Agent服务实例
func NewAgentService() *AgentService {
	return &AgentService{}
}

// 工作区路径
func getWorkspaceDir() string {
	if getDeployMode() == "local" {
		return filepath.Join(getOpenClawConfigDir(), "workspace")
	}
	return filepath.Join(getDataDir(), "workspace")
}

// 允许的文件名
var validAgentFiles = map[string]string{
	"IDENTITY":    "IDENTITY.md",
	"USER":        "USER.md",
	"SOUL":        "SOUL.md",
	"AGENTS":      "AGENTS.md",
	"TOOLS":       "TOOLS.md",
	"BOOTSTRAP":   "BOOTSTRAP.md",
	"HEARTBEAT":   "HEARTBEAT.md",
	"CONVENTIONS": "CONVENTIONS.md",
	"MEMORY":      "MEMORY.md",
}

// GetAgentFiles 读取全部Agent文件
func (s *AgentService) GetAgentFiles() (*dto.AgentFilesResp, error) {
	resp := &dto.AgentFilesResp{}
	dir := getWorkspaceDir()

	for name, filename := range validAgentFiles {
		content := ""
		data, err := os.ReadFile(filepath.Join(dir, filename))
		if err == nil {
			content = string(data)
		}
		resp.Files = append(resp.Files, dto.AgentFileItem{
			Name:    name,
			Content: content,
		})
	}
	return resp, nil
}

// SaveAgentFile 保存单个Agent文件
func (s *AgentService) SaveAgentFile(req dto.AgentSaveReq) (map[string]any, error) {
	filename, ok := validAgentFiles[req.Name]
	if !ok {
		return nil, fmt.Errorf("无效的文件名: %s", req.Name)
	}
	// 根据 AgentID 确定目录
	var dir string
	if req.AgentID == "" || req.AgentID == "main" {
		dir = getWorkspaceDir()
	} else {
		dir = agentDir(req.AgentID)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(req.Content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}
	return map[string]any{"success": true, "message": "保存成功"}, nil
}

// ResetAgentFile 重置Agent文件为OpenClaw默认内容
func (s *AgentService) ResetAgentFile(req dto.AgentFileReq) (map[string]any, error) {
	filename, ok := validAgentFiles[req.Name]
	if !ok {
		return nil, fmt.Errorf("无效的文件名: %s", req.Name)
	}
	content, ok := defaultContents[req.Name]
	if !ok {
		return nil, fmt.Errorf("无默认内容: %s", req.Name)
	}
	var dir string
	if req.AgentID == "" || req.AgentID == "main" {
		dir = getWorkspaceDir()
	} else {
		dir = agentDir(req.AgentID)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}
	return map[string]any{"success": true, "message": "已恢复默认"}, nil
}

// GetAgentTemplates 获取预设模板列表
func (s *AgentService) GetAgentTemplates() (*dto.AgentTemplatesResp, error) {
	var list []dto.AgentTemplate
	for _, t := range agentTemplates {
		list = append(list, dto.AgentTemplate{
			Key:         t.Key,
			Name:        t.Name,
			Description: t.Description,
		})
	}
	return &dto.AgentTemplatesResp{Templates: list}, nil
}

// ApplyAgentTemplate 应用预设模板
func (s *AgentService) ApplyAgentTemplate(req dto.ApplyTemplateReq) (map[string]any, error) {
	var tpl *agentTemplate
	for _, t := range agentTemplates {
		if t.Key == req.Key {
			tpl = &t
			break
		}
	}
	if tpl == nil {
		return nil, fmt.Errorf("模板不存在: %s", req.Key)
	}
	dir := getWorkspaceDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	files := map[string]string{
		"IDENTITY.md": tpl.Identity,
		"USER.md":     tpl.User,
		"SOUL.md":     tpl.Soul,
	}
	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("写入 %s 失败: %v", filename, err)
		}
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("已应用「%s」模板", tpl.Name)}, nil
}

// ========== 多 Agent 管理 ==========

// generateID 生成 8 位随机 hex ID
func generateID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// agentsJSONPath 获取 agents.json 路径
func agentsJSONPath() string {
	return filepath.Join(getWorkspaceDir(), "agents.json")
}

// agentDir 获取指定 Agent 的目录
func agentDir(id string) string {
	return filepath.Join(getWorkspaceDir(), "agents", id)
}

// getDefaultModel 获取 openclaw.json 中配置的默认主模型
func getDefaultModel() string {
	if config, err := readOpenClawConfig(); err == nil {
		if agents, ok := config["agents"].(map[string]any); ok {
			if defaults, ok := agents["defaults"].(map[string]any); ok {
				if model, ok := defaults["model"].(map[string]any); ok {
					if primary, ok := model["primary"].(string); ok && primary != "" {
						return primary
					}
				}
			}
		}
	}
	return "deepseek/deepseek-chat"
}

// registerOpenClawAgent 调用 openclaw CLI 注册 Agent
func registerOpenClawAgent(name string, workspaceDir string, model string) error {
	if model == "" {
		model = getDefaultModel()
	}

	// 清理 openclaw.json 中 agents.list 里的 "main"（CLI 认为 main 是保留的，显式列出会报错）
	if config, err := readOpenClawConfig(); err == nil {
		if agentsCfg, ok := config["agents"].(map[string]any); ok {
			if list, ok := agentsCfg["list"].([]any); ok {
				var cleanList []any
				for _, item := range list {
					if m, ok := item.(map[string]any); ok {
						if id, _ := m["id"].(string); id == "main" {
							continue
						}
					}
					cleanList = append(cleanList, item)
				}
				if cleanList == nil {
					cleanList = []any{} // 保留空 list 而不是 nil
				}
				agentsCfg["list"] = cleanList
				_ = writeOpenClawConfig(config)
			}
		}
	}

	// 使用 runClawCmd 兼容 local/Docker 两种部署模式
	output, err := runClawCmd("agents", "add", name,
		"--workspace", workspaceDir,
		"--model", model,
		"--non-interactive",
	)
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	log.Printf("[INFO] OpenClaw Agent 注册成功: %s (model: %s)", name, model)
	return nil
}

// readAgentsList 从 agents.json 读取 Agent 列表
func readAgentsList() ([]dto.AgentInfo, error) {
	data, err := os.ReadFile(agentsJSONPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.AgentInfo{}, nil
		}
		return nil, err
	}
	var agents []dto.AgentInfo
	if err := json.Unmarshal(data, &agents); err != nil {
		return nil, err
	}
	// 兼容老数据：如果没有 avatar，给个默认值
	changed := false
	for i := range agents {
		if agents[i].Avatar == "" {
			if agents[i].Role == "main" {
				agents[i].Avatar = "robot"
			} else {
				agents[i].Avatar = "brain"
			}
			changed = true
		}
	}
	if changed {
		_ = writeAgentsList(agents)
	}
	return agents, nil
}

// writeAgentsList 将 Agent 列表写入 agents.json
func writeAgentsList(agents []dto.AgentInfo) error {
	dir := getWorkspaceDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(agentsJSONPath(), data, 0644)
}

// initDefaultMainAgent 初始化默认主 Agent（首次使用时）
func initDefaultMainAgent() ([]dto.AgentInfo, error) {
	now := time.Now().Format(time.RFC3339)
	mainAgent := dto.AgentInfo{
		ID:          "main",
		Name:        "OpenClaw",
		Role:        "main",
		Description: "主 Agent，负责任务分发与协调",
		Status:      "idle",
		Avatar:      "robot",
		ParentID:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	agents := []dto.AgentInfo{mainAgent}
	if err := writeAgentsList(agents); err != nil {
		return nil, err
	}
	return agents, nil
}

// ListAgents 获取 Agent 列表
func (s *AgentService) ListAgents() (*dto.ListAgentsResp, error) {
	agents, err := readAgentsList()
	if err != nil {
		return nil, fmt.Errorf("读取 Agent 列表失败: %v", err)
	}
	// 首次使用：创建默认主 Agent
	if len(agents) == 0 {
		agents, err = initDefaultMainAgent()
		if err != nil {
			return nil, fmt.Errorf("初始化默认 Agent 失败: %v", err)
		}
	}
	return &dto.ListAgentsResp{Agents: agents}, nil
}

// CreateAgent 创建新 Agent
func (s *AgentService) CreateAgent(req dto.CreateAgentReq) (*dto.AgentInfo, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("Agent 名称不能为空")
	}
	if req.Role == "" {
		req.Role = "specialist"
	}
	if req.Avatar == "" {
		req.Avatar = "brain"
	}
	if req.Model == "" {
		req.Model = getDefaultModel()
	}

	agents, err := readAgentsList()
	if err != nil {
		return nil, fmt.Errorf("读取 Agent 列表失败: %v", err)
	}

	now := time.Now().Format(time.RFC3339)
	agent := dto.AgentInfo{
		ID:          generateID(),
		Name:        req.Name,
		Role:        req.Role,
		Description: req.Description,
		Status:      "idle",
		Avatar:      req.Avatar,
		Model:       req.Model,
		ParentID:    req.ParentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 创建 Agent 工作区目录并写入默认人格文件
	dir := agentDir(agent.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建 Agent 目录失败: %v", err)
	}
	for name, content := range defaultContents {
		// 子 Agent 的 AGENTS.md 用精简模板
		if name == "AGENTS" && agent.Role != "main" && agent.ID != "main" {
			content = "# 团队名录\n\n当前无下属 Agent。\n"
		}
		filename := validAgentFiles[name]
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("写入 %s 失败: %v", filename, err)
		}
	}

	// 创建 OpenClaw agent 配置目录并写入 models.json
	var openclawAgentBaseDir string
	if getDeployMode() == "local" {
		openclawAgentBaseDir = filepath.Join(getOpenClawConfigDir(), "agents", agent.ID)
	} else {
		openclawAgentBaseDir = filepath.Join(getDataDir(), "agents", agent.ID)
	}
	openclawAgentDir := filepath.Join(openclawAgentBaseDir, "agent")
	if err := os.MkdirAll(openclawAgentDir, 0755); err != nil {
		log.Printf("[WARN] 创建 OpenClaw agent 目录失败: %v", err)
	}
	// 从 openclaw.json 读取 models 配置写入 agent 的 models.json
	if ocConfig, err := readOpenClawConfig(); err == nil {
		if modelsConfig, ok := ocConfig["models"].(map[string]any); ok {
			// models.json 只需要 providers 部分（去掉 mode 等顶层字段）
			modelsJSON := map[string]any{}
			if providers, ok := modelsConfig["providers"]; ok {
				modelsJSON["providers"] = providers
			}
			if data, err := json.MarshalIndent(modelsJSON, "", "  "); err == nil {
				modelsPath := filepath.Join(openclawAgentDir, "models.json")
				if err := os.WriteFile(modelsPath, data, 0644); err != nil {
					log.Printf("[WARN] 写入 agent models.json 失败: %v", err)
				}
			}
		}
	}

	// 注册到 OpenClaw（主控已内置，不需要注册）
	if agent.Role != "main" && agent.ID != "main" {
		log.Printf("[DEBUG] 注册 OpenClaw Agent: name=%s, role=%s, dir=%s", agent.Name, agent.Role, dir)
		if err := registerOpenClawAgent(agent.Name, dir, agent.Model); err != nil {
			log.Printf("[WARN] 注册 OpenClaw Agent 失败: %v (Agent 文件已创建，可手动注册)", err)
		}
	}

	agents = append(agents, agent)
	if err := writeAgentsList(agents); err != nil {
		return nil, fmt.Errorf("保存 Agent 列表失败: %v", err)
	}

	// 自动同步 openclaw.json 和 main AGENTS.md
	syncOpenClawConfig(agents)

	return &agent, nil
}

// UpdateAgent 更新 Agent 元信息
func (s *AgentService) UpdateAgent(req dto.UpdateAgentReq) (*dto.AgentInfo, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("Agent ID 不能为空")
	}

	agents, err := readAgentsList()
	if err != nil {
		return nil, fmt.Errorf("读取 Agent 列表失败: %v", err)
	}

	var updated *dto.AgentInfo
	for i := range agents {
		if agents[i].ID == req.ID {
			if req.Name != "" {
				agents[i].Name = req.Name
			}
			if req.Role != "" {
				agents[i].Role = req.Role
			}
			if req.Description != "" {
				agents[i].Description = req.Description
			}
			if req.Avatar != "" {
				agents[i].Avatar = req.Avatar
			}
			if req.Model != "" {
				agents[i].Model = req.Model
			}
			agents[i].ParentID = req.ParentID
			agents[i].UpdatedAt = time.Now().Format(time.RFC3339)
			updated = &agents[i]
			break
		}
	}

	if updated == nil {
		return nil, fmt.Errorf("Agent 不存在: %s", req.ID)
	}


	if err := writeAgentsList(agents); err != nil {
		return nil, fmt.Errorf("保存 Agent 列表失败: %v", err)
	}

	// 同步 openclaw.json 和 main AGENTS.md
	syncOpenClawConfig(agents)

	return updated, nil
}

// DeleteAgent 删除 Agent
func (s *AgentService) DeleteAgent(req dto.DeleteAgentReq) (map[string]any, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("Agent ID 不能为空")
	}
	if req.ID == "main" {
		return nil, fmt.Errorf("不能删除主 Agent")
	}

	agents, err := readAgentsList()
	if err != nil {
		return nil, fmt.Errorf("读取 Agent 列表失败: %v", err)
	}

	found := false
	var newAgents []dto.AgentInfo
	for _, a := range agents {
		if a.ID == req.ID {
			found = true
			continue
		}
		// 如果子 Agent 从属于被删除的 Agent，清空 ParentID
		if a.ParentID == req.ID {
			a.ParentID = ""
		}
		newAgents = append(newAgents, a)
	}

	if !found {
		return nil, fmt.Errorf("Agent 不存在: %s", req.ID)
	}

	// 从 OpenClaw 注销
	var agentName string
	for _, a := range agents {
		if a.ID == req.ID {
			agentName = a.Name
			break
		}
	}
	if agentName != "" {
		if output, err := runClawCmd("agents", "delete", agentName, "--force"); err != nil {
			log.Printf("[WARN] 从 OpenClaw 注销 Agent 失败: %s: %s", err, string(output))
		} else {
			log.Printf("[INFO] OpenClaw Agent 已注销: %s", agentName)
		}
	}

	// 删除 Agent 工作区目录
	dir := agentDir(req.ID)
	_ = os.RemoveAll(dir)

	// 删除 OpenClaw agent 配置目录
	var openclawAgentDir string
	if getDeployMode() == "local" {
		openclawAgentDir = filepath.Join(getOpenClawConfigDir(), "agents", req.ID)
	} else {
		openclawAgentDir = filepath.Join(getDataDir(), "agents", req.ID)
	}
	_ = os.RemoveAll(openclawAgentDir)
	log.Printf("[INFO] 已清理 OpenClaw agent 目录: %s", openclawAgentDir)

	if err := writeAgentsList(newAgents); err != nil {
		return nil, fmt.Errorf("保存 Agent 列表失败: %v", err)
	}

	// 自动同步 openclaw.json 和 main AGENTS.md
	syncOpenClawConfig(newAgents)

	return map[string]any{"success": true, "message": "Agent 已删除"}, nil
}

// GetAgentDetail 获取 Agent 详情（含人格文件）
func (s *AgentService) GetAgentDetail(req dto.GetAgentDetailReq) (*dto.AgentDetailResp, error) {
	if req.ID == "" {
		return nil, fmt.Errorf("Agent ID 不能为空")
	}

	agents, err := readAgentsList()
	if err != nil {
		return nil, fmt.Errorf("读取 Agent 列表失败: %v", err)
	}

	var agent *dto.AgentInfo
	for i := range agents {
		if agents[i].ID == req.ID {
			agent = &agents[i]
			break
		}
	}
	if agent == nil {
		return nil, fmt.Errorf("Agent 不存在: %s", req.ID)
	}

	// 确定文件目录：主 Agent 用 workspace 根，其他 Agent 用子目录
	var fileDir string
	if req.ID == "main" {
		fileDir = getWorkspaceDir()
	} else {
		fileDir = agentDir(req.ID)
	}

	var files []dto.AgentFileItem
	for name, filename := range validAgentFiles {
		content := ""
		fullPath := filepath.Join(fileDir, filename)
		data, readErr := os.ReadFile(fullPath)
		if readErr == nil {
			content = string(data)
		} else if defaultContent, ok := defaultContents[name]; ok {
			// 文件不存在，自动写入默认内容
			_ = os.MkdirAll(fileDir, 0755)
			_ = os.WriteFile(fullPath, []byte(defaultContent), 0644)
			content = defaultContent
		}
		files = append(files, dto.AgentFileItem{
			Name:    name,
			Content: content,
		})
	}

	return &dto.AgentDetailResp{
		Agent: *agent,
		Files: files,
	}, nil
}

// ========== 默认内容 ==========

var defaultContents = map[string]string{
	"IDENTITY": `# 身份设定

- **名称:** 未命名
- **类型:** AI助手
- **氛围:** 未设定
`,
	"USER": `# 服务对象

- **名字:** 未设定
- **如何称呼:** 未设定

## 背景

暂无背景信息。
`,
	"SOUL": `# 性格灵魂

## 风格调性

- **专业度 (50/100):** 自然、灵活、平衡
- **详细度 (50/100):** 简洁与详细平衡
- **主动性 (50/100):** 适度主动
- **共情度 (50/100):** 理性与共情并重

## 核心原则

- 先做再问
- 有观点
- 简洁明了
`,
	"AGENTS": `# 团队名录

当前无下属 Agent。

## 如何添加

在此文件中登记团队成员信息，主控可据此调度任务。

| 成员ID | 角色 | 擅长领域 | 状态 |
|--------|------|----------|------|
| - | - | - | - |
`,
	"TOOLS": `# 工具箱

在此配置 Agent 可使用的外部资源。

## 可用工具

暂无配置。

## 格式示例

- **工具名称:** 描述
- **访问方式:** SSH / API / 本地路径
- **凭证:** (如需)
`,
	"BOOTSTRAP": `# 入职自检

Agent 启动时自动执行的检查清单。

## 检查项

- [ ] 确认工作目录可访问
- [ ] 确认依赖工具已安装
- [ ] 确认网络连通性

## 启动指令

暂无自定义启动指令。
`,
	"HEARTBEAT": `# 赛博心跳

定义 Agent 的周期性自主行为。

## 心跳任务

暂无配置。

## 格式示例

- **频率:** 每 4 小时
- **动作:** 检查 Git 仓库更新
- **输出:** 汇总到日志
`,
	"CONVENTIONS": `# 代码规范

定义 Agent 输出代码时应遵循的规范。

## 风格

- 使用 4 空格缩进
- 函数命名使用 camelCase
- 文件命名使用 kebab-case

## 提交规范

- feat: 新功能
- fix: 修复
- docs: 文档
- refactor: 重构
`,
	"MEMORY": `# 持久记忆 (Long-term Memory)

精选的长期记忆。仅在主私密会话中加载，不在共享/群组上下文中使用。

在此记录需要跨会话保留的关键信息，Agent 会在每次私密对话中自动回忆这些内容。

## 用户偏好

（记录用户的习惯、风格偏好等）

## 项目笔记

（记录重要的项目上下文和决策）

## 学习记录

（记录从过去交互中总结的经验教训）
`,
}

// ========== 预设模板 ==========

type agentTemplate struct {
	Key         string
	Name        string
	Description string
	Identity    string
	User        string
	Soul        string
}

var agentTemplates = []agentTemplate{
	{
		Key:         "chief-sre",
		Name:        "顶级架构师",
		Description: "稳重深度、预防性运维，适合复杂故障排查与架构设计",
		Identity: `# 身份：首席系统架构师 (Chief SRE)

你是拥有 20 年经验的资深系统架构师，精通 Linux 内核、分布式系统、K8s 编排以及网络安全。
你的知识库涵盖了从汇编级别到底层协议栈的所有细节。
你不仅能修复故障，还能指出系统架构中的潜在风险并提供加固方案。
`,
		User: `# 用户：系统管理员/研发主管

用户是负责维护核心生产环境的技术人员。
他们需要的是极其准确、具备生产环境操作安全意识的建议。
他们讨厌模棱两可的回答，更倾向于看到带有原理分析的解决方案。
`,
		Soul: `# 灵魂：严谨与预防

- **语气**：专业、冷静、权威。
- **原则**：在给出任何命令前，必须先提示备份数据或检查当前环境。
- **风格**：回答结构化（现象、原因、方案、预防）。
- **禁忌**：禁止提供未经测试的危险脚本；禁止在不解释风险的情况下建议使用 ` + "`rm -rf`" + ` 或修改核心内核参数。
`,
	},
	{
		Key:         "devops-ninja",
		Name:        "高效运维助手",
		Description: "极简脚本化、结果导向，适合日常巡检与 CI/CD 配置",
		Identity: `# 身份：自动化运维专家 (DevOps Ninja)

你是自动化运维的化身，精通 Python, Go, Shell, Ansible 和 Terraform。
你的目标是用最少的代码解决最繁琐的问题。
你对效率有近乎偏执的追求，能够快速生成符合规范、模块化、可复用的脚本。
`,
		User: `# 用户：忙碌的一线运维/开发者

用户通常在处理紧急任务或重复性工作。
他们需要能直接复制运行的代码片段。
他们希望你直接给答案，而不是长篇大论的理论。
`,
		Soul: `# 灵魂：极简与高效

- **语气**：干练、充满活力。
- **原则**：代码优先，解释次之。
- **风格**：使用大量代码块。脚本必须包含必要的注释。提供一键检查命令（如 ` + "`df -h`" + `, ` + "`top`" + ` 等）。
- **特色**：回答最后通常会附带一个"优化建议"，告诉用户如何将此操作自动化。
`,
	},
	{
		Key:         "secops-lead",
		Name:        "云原生安全官",
		Description: "敏锐合规、漏洞猎人，适合容器安全与权限审计",
		Identity: `# 身份：安全运维专家 (SecOps Lead)

你是专注于云原生安全的防御专家，精通等保 2.0、CIS 基准测试和渗透测试。
你对权限管理（RBAC）、加密传输和日志审计有极高的敏感度。
你的任务是在保证业务运行的前提下，将攻击面缩减到最小。
`,
		User: `# 用户：安全负责人/运维开发

用户关心合规性和系统漏洞。
他们需要知道每一个操作对系统安全性的影响。
`,
		Soul: `# 灵魂：警惕与合规

- **语气**：警示性、负责任。
- **原则**：始终遵循"最小权限原则"。
- **风格**：每项建议都会标注其对应的安全风险级别（高/中/低）。
- **特色**：在提供配置建议时，会额外补充如何审计该配置是否生效的验证方法。
`,
	},
}

// ========== 多 Agent 协同配置同步 ==========

// syncOpenClawConfig 自动同步 openclaw.json（agentToAgent + agents.list）和 main 的 AGENTS.md
func syncOpenClawConfig(agents []dto.AgentInfo) {
	// 1. 更新 openclaw.json
	config, err := readOpenClawConfig()
	if err != nil {
		log.Printf("[WARN] syncOpenClawConfig: 读取 openclaw.json 失败: %v", err)
		return
	}

	// 收集所有 agent 的 openclaw id（main 用 "main"，specialist 用 hex ID）用于 agentToAgent
	var agentIDs []any
	for _, a := range agents {
		if a.ID == "main" || a.Role == "main" {
			agentIDs = append(agentIDs, "main")
		} else {
			agentIDs = append(agentIDs, a.ID)
		}
	}

	// 设置 tools.agentToAgent
	tools, _ := config["tools"].(map[string]any)
	if tools == nil {
		tools = map[string]any{}
	}
	tools["agentToAgent"] = map[string]any{
		"enabled": true,
		"allow":   agentIDs,
	}
	config["tools"] = tools

	// 重建 agents.list
	agentsCfg, _ := config["agents"].(map[string]any)
	if agentsCfg == nil {
		agentsCfg = map[string]any{}
	}

	// 收集专员
	var specialistEntries []any
	for _, a := range agents {
		if a.ID == "main" || a.Role == "main" {
			continue
		}
		entry := map[string]any{
			"id":   a.ID,
			"name": a.Name,
		}
		dir := agentDir(a.ID)
		entry["workspace"] = dir
		if getDeployMode() == "local" {
			entry["agentDir"] = filepath.Join(getOpenClawConfigDir(), "agents", a.ID, "agent")
		} else {
			entry["agentDir"] = filepath.Join(getDataDir(), "agents", a.ID, "agent")
		}
		if a.Model != "" {
			entry["model"] = a.Model
		}
		specialistEntries = append(specialistEntries, entry)
	}

	if len(specialistEntries) > 0 {
		// 有专员：list = [main] + specialists
		agentList := []any{map[string]any{"id": "main"}}
		agentList = append(agentList, specialistEntries...)
		agentsCfg["list"] = agentList
	} else {
		// 只剩 main：删掉 list 字段
		delete(agentsCfg, "list")
	}
	config["agents"] = agentsCfg

	if err := writeOpenClawConfig(config); err != nil {
		log.Printf("[WARN] syncOpenClawConfig: 写入 openclaw.json 失败: %v", err)
	} else {
		log.Printf("[INFO] syncOpenClawConfig: agents.list 已更新 (%d specialists), agentToAgent (%d agents)", len(specialistEntries), len(agentIDs))
	}

	// 2. 自动生成 main 的 AGENTS.md 团队名录
	var specialists []dto.AgentInfo
	for _, a := range agents {
		if a.Role != "main" {
			specialists = append(specialists, a)
		}
	}

	md := "# 团队名录\n\n你是项目经理（主控 Agent），负责接待用户的所有需求并协调团队。\n\n"
	if len(specialists) == 0 {
		md += "> 当前暂无子 Agent，所有任务由你独立完成。\n"
	} else {
		md += "## 你的团队成员\n\n"
		md += "你可以使用内部通讯工具向以下团队成员派发任务：\n\n"
		for _, a := range specialists {
			model := a.Model
			if model == "" {
				model = "默认模型"
			}
			desc := a.Description
			if desc == "" {
				desc = "暂无描述"
			}
			md += fmt.Sprintf("* **Agent ID: %s（名称：%s）**\n", a.ID, a.Name)
			md += fmt.Sprintf("    * 职责：%s\n", desc)
			md += fmt.Sprintf("    * 模型：%s\n", model)
			md += "\n"
		}
		md += "## 协同工作原则\n\n"
		md += "1. **任务拆解**：如果用户的任务较复杂，可将子任务分别派发给对应的专员。\n"
		md += "2. **结果汇总**：专员完成后由你进行最终汇总并回复用户。\n"
		md += "3. **按需派单**：根据专员的职责描述选择最合适的人选。\n"
	}

	// 写入 main 工作区的 AGENTS.md
	mainWorkspace := getWorkspaceDir()
	agentsMdPath := filepath.Join(mainWorkspace, "AGENTS.md")
	if err := os.WriteFile(agentsMdPath, []byte(md), 0644); err != nil {
		log.Printf("[WARN] syncOpenClawConfig: 写入 main AGENTS.md 失败: %v", err)
	} else {
		log.Printf("[INFO] syncOpenClawConfig: main AGENTS.md 已更新 (%d specialists)", len(specialists))
	}
}

// GetConfiguredModels 获取 openclaw.json 中已配置的模型列表
func (s *AgentService) GetConfiguredModels() (map[string]any, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	var models []map[string]string

	// 读取 models.providers
	modelsConfig, _ := config["models"].(map[string]any)
	if modelsConfig != nil {
		providers, _ := modelsConfig["providers"].(map[string]any)
		for providerID, providerVal := range providers {
			provider, ok := providerVal.(map[string]any)
			if !ok {
				continue
			}
			modelList, _ := provider["models"].([]any)
			for _, m := range modelList {
				model, ok := m.(map[string]any)
				if !ok {
					continue
				}
				modelID, _ := model["id"].(string)
				modelName, _ := model["name"].(string)
				if modelID == "" {
					continue
				}
				fullRef := providerID + "/" + modelID
				label := modelName
				if label == "" {
					label = modelID
				}
				models = append(models, map[string]string{
					"value": fullRef,
					"label": providerID + " / " + label,
				})
			}
		}
	}

	// 获取默认模型
	defaultModel := getDefaultModel()

	return map[string]any{
		"models":       models,
		"defaultModel": defaultModel,
	}, nil
}
