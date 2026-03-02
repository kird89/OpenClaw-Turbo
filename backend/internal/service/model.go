package service

import (
	"encoding/json"
	"fmt"
	"log"
)

// ModelService 模型配置管理服务
type ModelService struct{}

func NewModelService() *ModelService { return &ModelService{} }

// GetModelsConfig 读取 openclaw.json 中的 models 配置 + 默认模型
func (s *ModelService) GetModelsConfig() (map[string]any, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	modelsConfig, _ := config["models"].(map[string]any)
	if modelsConfig == nil {
		modelsConfig = map[string]any{
			"mode":      "merge",
			"providers": map[string]any{},
		}
	}

	// 读取默认模型 agents.defaults.model.primary
	defaultModel := ""
	if agents, ok := config["agents"].(map[string]any); ok {
		if defaults, ok := agents["defaults"].(map[string]any); ok {
			if model, ok := defaults["model"].(map[string]any); ok {
				if primary, ok := model["primary"].(string); ok {
					defaultModel = primary
				}
			}
		}
	}

	return map[string]any{
		"models":       modelsConfig,
		"defaultModel": defaultModel,
	}, nil
}

// SaveModelsConfig 保存 models 配置 + 默认模型到 openclaw.json
func (s *ModelService) SaveModelsConfig(req map[string]any) (map[string]any, error) {
	modelsData, ok := req["models"]
	if !ok {
		return nil, fmt.Errorf("缺少 models 参数")
	}

	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	// 合并 models 配置（保留已有模型的元数据字段）
	mergedModels := mergeModelsConfig(config["models"], modelsData)
	config["models"] = mergedModels

	// 保存默认模型到 agents.defaults.model.primary
	if dm, ok := req["defaultModel"].(string); ok && dm != "" {
		agents, _ := config["agents"].(map[string]any)
		if agents == nil {
			agents = map[string]any{}
		}
		defaults, _ := agents["defaults"].(map[string]any)
		if defaults == nil {
			defaults = map[string]any{}
		}
		model, _ := defaults["model"].(map[string]any)
		if model == nil {
			model = map[string]any{}
		}
		model["primary"] = dm
		defaults["model"] = model
		agents["defaults"] = defaults
		config["agents"] = agents
	}

	if err := writeOpenClawConfig(config); err != nil {
		return nil, fmt.Errorf("保存配置失败: %v", err)
	}

	if data, err := json.MarshalIndent(modelsData, "", "  "); err == nil {
		log.Printf("[INFO] 模型配置已保存: %s", string(data))
	}

	return map[string]any{"success": true, "message": "模型配置已保存"}, nil
}

// mergeModelsConfig 合并新旧模型配置，保留已有模型的元数据字段
func mergeModelsConfig(existing, incoming any) map[string]any {
	existingMap, _ := existing.(map[string]any)
	incomingMap, _ := incoming.(map[string]any)
	if incomingMap == nil {
		return map[string]any{}
	}

	result := map[string]any{}
	// 复制 incoming 的顶层字段（如 mode）
	for k, v := range incomingMap {
		if k != "providers" {
			result[k] = v
		}
	}

	existingProviders := func() map[string]any {
		if existingMap == nil {
			return nil
		}
		p, _ := existingMap["providers"].(map[string]any)
		return p
	}()

	incomingProviders, _ := incomingMap["providers"].(map[string]any)
	mergedProviders := map[string]any{}

	for pid, pData := range incomingProviders {
		inProv, _ := pData.(map[string]any)
		if inProv == nil {
			continue
		}

		mergedProv := map[string]any{}
		// 复制供应商级别字段
		for k, v := range inProv {
			if k != "models" {
				mergedProv[k] = v
			}
		}

		// 获取已有的模型列表，建立 id -> model 索引
		existingModelsIndex := map[string]map[string]any{}
		if existingProviders != nil {
			if exProv, ok := existingProviders[pid].(map[string]any); ok {
				if exModels, ok := exProv["models"].([]any); ok {
					for _, m := range exModels {
						if em, ok := m.(map[string]any); ok {
							if id, ok := em["id"].(string); ok {
								existingModelsIndex[id] = em
							}
						}
					}
				}
			}
		}

		// 合并模型：以 incoming 为准，但保留已有的元数据字段
		inModels, _ := inProv["models"].([]any)
		mergedModels := []any{}
		for _, m := range inModels {
			inModel, _ := m.(map[string]any)
			if inModel == nil {
				continue
			}
			id, _ := inModel["id"].(string)
			if id == "" {
				continue
			}

			// 以已有模型为基础，覆盖 incoming 的字段
			merged := map[string]any{}
			if existing, ok := existingModelsIndex[id]; ok {
				// 已有模型：保留元数据
				for k, v := range existing {
					merged[k] = v
				}
			} else {
				// 新模型：套用默认元数据模板（与 deploy.go 初始化逻辑一致）
				merged = map[string]any{
					"contextWindow": 128000,
					"cost": map[string]any{
						"cacheRead": 0, "cacheWrite": 0,
						"input": 0, "output": 0,
					},
					"maxTokens": 8192,
					"reasoning": false,
				}
			}
			for k, v := range inModel {
				merged[k] = v
			}
			// 确保 name 字段存在
			if _, ok := merged["name"]; !ok {
				merged["name"] = id
			}
			mergedModels = append(mergedModels, merged)
		}
		mergedProv["models"] = mergedModels
		mergedProviders[pid] = mergedProv
	}

	result["providers"] = mergedProviders
	return result
}
