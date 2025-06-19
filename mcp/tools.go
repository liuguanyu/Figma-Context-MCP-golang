package mcp

import (
	"fmt"

	"figma-mcp-server/figma"
	"figma-mcp-server/types"
)

// GetAvailableTools 返回可用的工具列表
func GetAvailableTools() []types.Tool {
	return []types.Tool{
		{
			Name:        "get_figma_data",
			Description: "获取Figma文件的布局信息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"figmaApiKey": map[string]interface{}{
						"type":        "string",
						"description": "Figma API认证密钥",
					},
					"fileKey": map[string]interface{}{
						"type":        "string",
						"description": "Figma文件ID",
					},
					"nodeId": map[string]interface{}{
						"type":        "string",
						"description": "特定节点ID",
					},
					"depth": map[string]interface{}{
						"type":        "number",
						"description": "遍历深度",
					},
				},
				"required": []string{"figmaApiKey", "fileKey"},
			},
		},
		{
			Name:        "download_figma_images",
			Description: "下载Figma文件中的SVG/PNG图像",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"figmaApiKey": map[string]interface{}{
						"type":        "string",
						"description": "Figma API认证密钥",
					},
					"fileKey": map[string]interface{}{
						"type":        "string",
						"description": "Figma文件ID",
					},
					"nodes": map[string]interface{}{
						"type":        "array",
						"description": "包含nodeId、fileName等的节点数组",
					},
					"localPath": map[string]interface{}{
						"type":        "string",
						"description": "本地存储路径",
					},
					"pngScale": map[string]interface{}{
						"type":        "number",
						"description": "PNG缩放比例",
					},
					"svgOptions": map[string]interface{}{
						"type":        "object",
						"description": "SVG导出选项",
					},
				},
				"required": []string{"figmaApiKey", "fileKey", "nodes", "localPath"},
			},
		},
	}
}

// CallTool 调用指定的工具
func CallTool(toolName string, arguments map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "get_figma_data":
		return callGetFigmaData(arguments)
	case "download_figma_images":
		return callDownloadFigmaImages(arguments)
	default:
		return nil, fmt.Errorf("未知工具: %s", toolName)
	}
}

func callGetFigmaData(args map[string]interface{}) (interface{}, error) {
	// 提取参数
	figmaApiKey, ok := args["figmaApiKey"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: figmaApiKey")
	}

	fileKey, ok := args["fileKey"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: fileKey")
	}

	nodeId, _ := args["nodeId"].(string)

	var depth int
	if d, ok := args["depth"].(float64); ok {
		depth = int(d)
	}

	// 调用Figma服务
	result, err := figma.GetFigmaData(figmaApiKey, fileKey, nodeId, depth)
	if err != nil {
		return types.ToolResult{
			Content: []types.Content{{
				Type: "text",
				Text: fmt.Sprintf("错误: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return types.ToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("成功获取Figma数据: %v", result),
		}},
	}, nil
}

func callDownloadFigmaImages(args map[string]interface{}) (interface{}, error) {
	// 提取参数
	figmaApiKey, ok := args["figmaApiKey"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: figmaApiKey")
	}

	fileKey, ok := args["fileKey"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: fileKey")
	}

	nodes, ok := args["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: nodes")
	}

	localPath, ok := args["localPath"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少必需参数: localPath")
	}

	pngScale := 1.0
	if scale, ok := args["pngScale"].(float64); ok {
		pngScale = scale
	}

	svgOptions := map[string]interface{}{}
	if opts, ok := args["svgOptions"].(map[string]interface{}); ok {
		svgOptions = opts
	}

	// 调用Figma服务
	err := figma.DownloadFigmaImages(figmaApiKey, fileKey, nodes, localPath, pngScale, svgOptions)
	if err != nil {
		return types.ToolResult{
			Content: []types.Content{{
				Type: "text",
				Text: fmt.Sprintf("错误: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return types.ToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: "成功下载图像",
		}},
	}, nil
}
