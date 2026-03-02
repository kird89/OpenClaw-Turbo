package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvVarService 环境变量管理服务
type EnvVarService struct{}

func NewEnvVarService() *EnvVarService { return &EnvVarService{} }

// pluginAnchor 插件环境变量分隔锚点
const pluginAnchor = "# === PLUGINS ==="

// getEnvFilePath 获取 .env 文件路径
func getEnvFilePath() string {
	if getDeployMode() == "local" {
		return filepath.Join(getOpenClawConfigDir(), ".env")
	}
	return filepath.Join(getDataDir(), ".env")
}

// ListEnvVars 读取 .env 文件中锚点以下的 plugin 环境变量
func (s *EnvVarService) ListEnvVars() (map[string]any, error) {
	envPath := getEnvFilePath()
	vars := []map[string]string{}

	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"vars": vars}, nil
		}
		return nil, fmt.Errorf("读取 .env 文件失败: %v", err)
	}

	// 只读取锚点之后的变量
	inPluginSection := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == pluginAnchor {
			inPluginSection = true
			continue
		}
		if !inPluginSection {
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])
		vars = append(vars, map[string]string{"key": key, "value": val})
	}

	return map[string]any{"vars": vars}, nil
}

// SaveEnvVars 保存插件环境变量（只修改锚点之后的部分）
func (s *EnvVarService) SaveEnvVars(req map[string]any) (map[string]any, error) {
	rawVars, ok := req["vars"]
	if !ok {
		return nil, fmt.Errorf("缺少 vars 参数")
	}

	// 解析传入的 vars 数组
	newVars := []struct{ key, val string }{}
	varList, ok := rawVars.([]any)
	if !ok {
		return nil, fmt.Errorf("vars 参数格式错误")
	}
	for _, item := range varList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key, _ := m["key"].(string)
		val, _ := m["value"].(string)
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		newVars = append(newVars, struct{ key, val string }{key, val})
	}

	envPath := getEnvFilePath()
	os.MkdirAll(filepath.Dir(envPath), 0755)

	// 读取现有文件，保留锚点之前的系统部分
	var systemLines []string
	data, err := os.ReadFile(envPath)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == pluginAnchor {
				break // 到锚点就停，后面的都丢弃
			}
			systemLines = append(systemLines, line)
		}
	}

	// 去掉系统部分尾部空行
	for len(systemLines) > 0 && strings.TrimSpace(systemLines[len(systemLines)-1]) == "" {
		systemLines = systemLines[:len(systemLines)-1]
	}

	// 重建文件: 系统部分 + 空行 + 锚点 + 插件变量
	var outLines []string
	outLines = append(outLines, systemLines...)
	outLines = append(outLines, "", pluginAnchor)
	for _, v := range newVars {
		outLines = append(outLines, v.key+"="+v.val)
	}
	outLines = append(outLines, "") // 文件末尾换行

	if err := os.WriteFile(envPath, []byte(strings.Join(outLines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("写入 .env 文件失败: %v", err)
	}

	return map[string]any{"success": true, "path": envPath}, nil
}
