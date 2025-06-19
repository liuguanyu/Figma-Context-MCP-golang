package figma

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"figma-mcp-server/types"

	"gopkg.in/yaml.v2"
)

var (
	// 简化的HTTP客户端
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// GetFigmaData 获取Figma文件数据，简化版本
func GetFigmaData(figmaApiKey, fileKey, nodeId string, depth int) (string, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey)
	if nodeId != "" {
		url += fmt.Sprintf("/nodes?ids=%s", nodeId)
		if depth > 0 {
			url += fmt.Sprintf("&depth=%d", depth)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-FIGMA-TOKEN", figmaApiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Figma API错误: %d", resp.StatusCode)
	}

	var simplifiedDesign *types.SimplifiedDesign
	if nodeId != "" {
		simplifiedDesign, err = parseFigmaNodeResponse(resp.Body)
	} else {
		simplifiedDesign, err = parseFigmaFileResponse(resp.Body)
	}

	if err != nil {
		return "", err
	}

	// 构建结果结构
	result := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":         simplifiedDesign.Name,
			"lastModified": simplifiedDesign.LastModified,
			"thumbnailUrl": simplifiedDesign.ThumbnailUrl,
		},
		"components":    simplifiedDesign.Components,
		"componentSets": simplifiedDesign.ComponentSets,
		"nodes":         simplifiedDesign.Nodes,
		"globalVars":    simplifiedDesign.GlobalVars,
	}

	yamlData, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

// 简化的文件响应解析
func parseFigmaFileResponse(body io.Reader) (*types.SimplifiedDesign, error) {
	var apiResponse types.FigmaAPIResponse
	if err := json.NewDecoder(body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	simplifiedDesign := &types.SimplifiedDesign{
		Name:          apiResponse.Name,
		LastModified:  apiResponse.LastModified,
		ThumbnailUrl:  apiResponse.ThumbnailUrl,
		Components:    apiResponse.Components,
		ComponentSets: apiResponse.ComponentSets,
		GlobalVars:    types.GlobalVars{Styles: make(map[string]interface{})},
	}

	// 简单处理子节点
	for _, child := range apiResponse.Document.Children {
		if isVisible(child) {
			if node := parseNode(child, simplifiedDesign.GlobalVars.Styles, nil); node != nil {
				simplifiedDesign.Nodes = append(simplifiedDesign.Nodes, *node)
			}
		}
	}

	return simplifiedDesign, nil
}

// 简化的节点响应解析
func parseFigmaNodeResponse(body io.Reader) (*types.SimplifiedDesign, error) {
	var apiResponse types.FigmaAPINodeResponse
	if err := json.NewDecoder(body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	simplifiedDesign := &types.SimplifiedDesign{
		Name:          apiResponse.Name,
		LastModified:  apiResponse.LastModified,
		ThumbnailUrl:  apiResponse.ThumbnailUrl,
		Components:    make(map[string]interface{}),
		ComponentSets: make(map[string]interface{}),
		GlobalVars:    types.GlobalVars{Styles: make(map[string]interface{})},
	}

	// 合并组件
	for _, nodeWrapper := range apiResponse.Nodes {
		if nodeWrapper.Components != nil {
			for k, v := range nodeWrapper.Components {
				simplifiedDesign.Components[k] = v
			}
		}
		if nodeWrapper.ComponentSets != nil {
			for k, v := range nodeWrapper.ComponentSets {
				simplifiedDesign.ComponentSets[k] = v
			}
		}
	}

	// 解析节点
	for _, nodeWrapper := range apiResponse.Nodes {
		if isVisible(nodeWrapper.Document) {
			if node := parseNode(nodeWrapper.Document, simplifiedDesign.GlobalVars.Styles, nil); node != nil {
				simplifiedDesign.Nodes = append(simplifiedDesign.Nodes, *node)
			}
		}
	}

	return simplifiedDesign, nil
}

// 简化的节点解析
func parseNode(figmaNode types.FigmaNode, globalStyles map[string]interface{}, parent *types.FigmaNode) *types.SimplifiedNode {
	simplified := &types.SimplifiedNode{
		ID:   figmaNode.ID,
		Name: figmaNode.Name,
		Type: figmaNode.Type,
	}

	// 处理文本
	if figmaNode.Characters != "" {
		simplified.Text = figmaNode.Characters
	}

	// 处理样式
	if len(figmaNode.Style) > 0 {
		simplified.TextStyle = findOrCreateVar(globalStyles, figmaNode.Style, "style")
	}

	if len(figmaNode.Fills) > 0 {
		simplified.Fills = findOrCreateVar(globalStyles, figmaNode.Fills, "fill")
	}

	if len(figmaNode.Strokes) > 0 || figmaNode.StrokeWeight > 0 {
		strokeData := map[string]interface{}{
			"colors": figmaNode.Strokes,
			"weight": figmaNode.StrokeWeight,
			"align":  figmaNode.StrokeAlign,
		}
		simplified.Strokes = findOrCreateVar(globalStyles, strokeData, "stroke")
	}

	if len(figmaNode.Effects) > 0 {
		simplified.Effects = findOrCreateVar(globalStyles, figmaNode.Effects, "effect")
	}

	// 处理透明度
	if figmaNode.Opacity != nil && *figmaNode.Opacity != 1.0 {
		simplified.Opacity = figmaNode.Opacity
	}

	// 简化圆角处理
	if figmaNode.CornerRadius != nil {
		simplified.BorderRadius = fmt.Sprintf("%.0fpx", *figmaNode.CornerRadius)
	} else if len(figmaNode.RectangleCornerRadii) == 4 {
		simplified.BorderRadius = fmt.Sprintf("%.0fpx %.0fpx %.0fpx %.0fpx",
			figmaNode.RectangleCornerRadii[0], figmaNode.RectangleCornerRadii[1],
			figmaNode.RectangleCornerRadii[2], figmaNode.RectangleCornerRadii[3])
	}

	// 处理layout信息
	if hasLayoutProperties(figmaNode) {
		layoutData := buildLayoutData(figmaNode, parent)
		if layoutJson, err := json.Marshal(layoutData); err == nil {
			simplified.Layout = string(layoutJson)
		}
	}

	// 处理组件ID
	if figmaNode.ComponentId != "" {
		simplified.ComponentId = figmaNode.ComponentId
	}

	// 处理组件属性
	if len(figmaNode.ComponentProperties) > 0 {
		for name, prop := range figmaNode.ComponentProperties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				componentProp := types.ComponentProperty{Name: name}
				if value, exists := propMap["value"]; exists {
					componentProp.Value = fmt.Sprintf("%v", value)
				}
				if propType, exists := propMap["type"]; exists {
					componentProp.Type = fmt.Sprintf("%v", propType)
				}
				simplified.ComponentProperties = append(simplified.ComponentProperties, componentProp)
			}
		}
	}

	// 递归处理子节点
	if len(figmaNode.Children) > 0 {
		for _, child := range figmaNode.Children {
			if isVisible(child) {
				if childNode := parseNode(child, globalStyles, &figmaNode); childNode != nil {
					simplified.Children = append(simplified.Children, *childNode)
				}
			}
		}
	}

	// 转换VECTOR为IMAGE-SVG
	if simplified.Type == "VECTOR" {
		simplified.Type = "IMAGE-SVG"
	}

	return simplified
}

// 简化的layout数据构建
func buildLayoutData(node types.FigmaNode, parent *types.FigmaNode) map[string]interface{} {
	layout := make(map[string]interface{})

	// 设置mode
	if node.LayoutMode != "" {
		layout["mode"] = strings.ToLower(node.LayoutMode)
	} else {
		layout["mode"] = "none"
	}

	// 对齐属性
	if node.PrimaryAxisAlignItems != "" {
		switch node.PrimaryAxisAlignItems {
		case "MIN":
			layout["justifyContent"] = "flex-start"
		case "CENTER":
			layout["justifyContent"] = "center"
		case "MAX":
			layout["justifyContent"] = "flex-end"
		case "SPACE_BETWEEN":
			layout["justifyContent"] = "space-between"
		}
	}

	if node.CounterAxisAlignItems != "" {
		switch node.CounterAxisAlignItems {
		case "MIN":
			layout["alignItems"] = "flex-start"
		case "CENTER":
			layout["alignItems"] = "center"
		case "MAX":
			layout["alignItems"] = "flex-end"
		case "BASELINE":
			layout["alignItems"] = "baseline"
		}
	}

	if node.LayoutAlign != "" {
		switch node.LayoutAlign {
		case "INHERIT":
			layout["alignSelf"] = "auto"
		case "MIN":
			layout["alignSelf"] = "flex-start"
		case "CENTER":
			layout["alignSelf"] = "center"
		case "MAX":
			layout["alignSelf"] = "flex-end"
		case "STRETCH":
			layout["alignSelf"] = "stretch"
		}
	}

	if node.LayoutWrap != "" {
		layout["wrap"] = node.LayoutWrap == "WRAP"
	}

	if node.ItemSpacing != nil {
		layout["gap"] = fmt.Sprintf("%.0fpx", *node.ItemSpacing)
	}

	// bbox和尺寸
	if node.AbsoluteBoundingBox != nil {
		layout["locationRelativeToParent"] = map[string]interface{}{
			"x": node.AbsoluteBoundingBox.X,
			"y": node.AbsoluteBoundingBox.Y,
		}
		layout["dimensions"] = map[string]interface{}{
			"width":  node.AbsoluteBoundingBox.Width,
			"height": node.AbsoluteBoundingBox.Height,
		}
	}

	// padding处理
	if node.PaddingTop != nil || node.PaddingRight != nil || node.PaddingBottom != nil || node.PaddingLeft != nil {
		if node.PaddingTop != nil && node.PaddingRight != nil && node.PaddingBottom != nil && node.PaddingLeft != nil {
			if *node.PaddingTop == *node.PaddingRight && *node.PaddingRight == *node.PaddingBottom && *node.PaddingBottom == *node.PaddingLeft {
				layout["padding"] = fmt.Sprintf("%.0fpx", *node.PaddingTop)
			} else {
				layout["padding"] = fmt.Sprintf("%.0fpx %.0fpx %.0fpx %.0fpx",
					*node.PaddingTop, *node.PaddingRight, *node.PaddingBottom, *node.PaddingLeft)
			}
		}
	}

	// sizing
	if node.LayoutSizingHorizontal != "" || node.LayoutSizingVertical != "" {
		sizing := make(map[string]interface{})
		if node.LayoutSizingHorizontal != "" {
			sizing["horizontal"] = strings.ToLower(node.LayoutSizingHorizontal)
		}
		if node.LayoutSizingVertical != "" {
			sizing["vertical"] = strings.ToLower(node.LayoutSizingVertical)
		}
		layout["sizing"] = sizing
	}

	if len(node.OverflowDirection) > 0 {
		var scrollDirs []string
		for _, dir := range node.OverflowDirection {
			scrollDirs = append(scrollDirs, strings.ToLower(dir))
		}
		layout["overflowScroll"] = scrollDirs
	}

	if node.LayoutPositioning != "" {
		switch node.LayoutPositioning {
		case "AUTO":
			layout["position"] = "relative"
		case "ABSOLUTE":
			layout["position"] = "absolute"
		}
	}

	return layout
}

// hasLayoutProperties 检查节点是否有layout属性
func hasLayoutProperties(node types.FigmaNode) bool {
	return node.LayoutMode != "" ||
		node.PrimaryAxisAlignItems != "" ||
		node.CounterAxisAlignItems != "" ||
		node.LayoutAlign != "" ||
		node.LayoutWrap != "" ||
		node.ItemSpacing != nil ||
		node.PaddingTop != nil ||
		node.PaddingRight != nil ||
		node.PaddingBottom != nil ||
		node.PaddingLeft != nil ||
		node.LayoutSizingHorizontal != "" ||
		node.LayoutSizingVertical != "" ||
		node.LayoutPositioning != "" ||
		node.LayoutGrow != nil ||
		len(node.OverflowDirection) > 0 ||
		node.AbsoluteBoundingBox != nil
}

// isVisible 检查节点是否可见
func isVisible(node types.FigmaNode) bool {
	if node.Visible != nil {
		return *node.Visible
	}
	return true
}

// 简化的变量查找/创建
func findOrCreateVar(globalStyles map[string]interface{}, value interface{}, prefix string) string {
	// 检查是否已存在相同的值
	valueJson, _ := json.Marshal(value)
	valueStr := string(valueJson)

	for existingId, existingValue := range globalStyles {
		existingJson, _ := json.Marshal(existingValue)
		if string(existingJson) == valueStr {
			return existingId
		}
	}

	// 创建新的变量ID
	varId := generateVarId(globalStyles, prefix)
	globalStyles[varId] = value
	return varId
}

// 简化的变量ID生成
func generateVarId(globalStyles map[string]interface{}, prefix string) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	for {
		id := prefix + "_"
		for i := 0; i < 6; i++ {
			id += string(chars[rand.Intn(len(chars))])
		}

		if _, exists := globalStyles[id]; !exists {
			return id
		}
	}
}

// DownloadFigmaImages 简化的图片下载
func DownloadFigmaImages(figmaApiKey, fileKey string, nodes []interface{}, localPath string, pngScale float64, svgOptions map[string]interface{}) error {
	// 解析节点列表
	var imageNodes []types.ImageNode
	for _, node := range nodes {
		if nodeMap, ok := node.(map[string]interface{}); ok {
			imageNode := types.ImageNode{}

			if nodeId, exists := nodeMap["nodeId"].(string); exists {
				imageNode.NodeId = nodeId
			} else {
				continue
			}

			if imageRef, exists := nodeMap["imageRef"].(string); exists {
				imageNode.ImageRef = imageRef
			}

			if fileName, exists := nodeMap["fileName"].(string); exists {
				imageNode.FileName = fileName
			} else {
				continue
			}

			imageNodes = append(imageNodes, imageNode)
		}
	}

	if len(imageNodes) == 0 {
		return fmt.Errorf("没有有效的图像节点")
	}

	// 创建本地目录
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 分离不同类型的节点
	var svgNodes, pngNodes, imageRefNodes []types.ImageNode

	for _, node := range imageNodes {
		if node.ImageRef != "" {
			imageRefNodes = append(imageRefNodes, node)
		} else if strings.HasSuffix(strings.ToLower(node.FileName), ".svg") {
			svgNodes = append(svgNodes, node)
		} else {
			pngNodes = append(pngNodes, node)
		}
	}

	// 简单的顺序下载
	if len(svgNodes) > 0 {
		if err := downloadImages(figmaApiKey, fileKey, svgNodes, localPath, "svg", svgOptions, 1.0); err != nil {
			return fmt.Errorf("下载SVG图像失败: %v", err)
		}
	}

	if len(pngNodes) > 0 {
		if err := downloadImages(figmaApiKey, fileKey, pngNodes, localPath, "png", nil, pngScale); err != nil {
			return fmt.Errorf("下载PNG图像失败: %v", err)
		}
	}

	if len(imageRefNodes) > 0 {
		if err := downloadImageFills(figmaApiKey, fileKey, imageRefNodes, localPath); err != nil {
			return fmt.Errorf("下载ImageRef图像失败: %v", err)
		}
	}

	fmt.Printf("成功下载 %d 个图像到: %s\n", len(imageNodes), localPath)
	return nil
}

// downloadImages 简化的图片下载
func downloadImages(figmaApiKey, fileKey string, nodes []types.ImageNode, localPath, format string, options map[string]interface{}, scale float64) error {
	var nodeIds []string
	for _, node := range nodes {
		nodeIds = append(nodeIds, node.NodeId)
	}

	apiURL := fmt.Sprintf("https://api.figma.com/v1/images/%s", fileKey)
	params := url.Values{}
	params.Add("ids", strings.Join(nodeIds, ","))
	params.Add("format", format)

	if format == "png" && scale != 1.0 {
		params.Add("scale", strconv.FormatFloat(scale, 'f', 1, 64))
	}

	if format == "svg" && options != nil {
		if outlineText, ok := options["outlineText"].(bool); ok && outlineText {
			params.Add("svg_outline_text", "true")
		}
		if includeId, ok := options["includeId"].(bool); ok && includeId {
			params.Add("svg_include_id", "true")
		}
		if simplifyStroke, ok := options["simplifyStroke"].(bool); ok && simplifyStroke {
			params.Add("svg_simplify_stroke", "true")
		}
	}

	fullURL := apiURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-FIGMA-TOKEN", figmaApiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Figma API错误: %d", resp.StatusCode)
	}

	var imagesResp types.FigmaImagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&imagesResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if imagesResp.Error != "" {
		return fmt.Errorf("Figma API error: %s", imagesResp.Error)
	}

	// 顺序下载文件
	for _, node := range nodes {
		imageURL, exists := imagesResp.Images[node.NodeId]
		if !exists || imageURL == "" {
			fmt.Printf("警告: 节点 %s 没有找到有效的图像URL\n", node.NodeId)
			continue
		}

		if err := downloadFile(imageURL, filepath.Join(localPath, node.FileName)); err != nil {
			fmt.Printf("警告: 下载文件 %s 失败: %v\n", node.FileName, err)
		} else {
			fmt.Printf("成功下载: %s\n", node.FileName)
		}
	}

	return nil
}

// downloadImageFills 简化的图像填充下载
func downloadImageFills(figmaApiKey, fileKey string, nodes []types.ImageNode, localPath string) error {
	apiURL := fmt.Sprintf("https://api.figma.com/v1/files/%s/images", fileKey)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-FIGMA-TOKEN", figmaApiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Figma API错误: %d", resp.StatusCode)
	}

	var fillsResp types.FigmaImageFillsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fillsResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if fillsResp.Error != "" {
		return fmt.Errorf("Figma API error: %s", fillsResp.Error)
	}

	// 顺序下载文件
	for _, node := range nodes {
		imageURL, exists := fillsResp.Meta.Images[node.ImageRef]
		if !exists || imageURL == "" {
			fmt.Printf("警告: ImageRef %s 没有找到有效的图像URL\n", node.ImageRef)
			continue
		}

		if err := downloadFile(imageURL, filepath.Join(localPath, node.FileName)); err != nil {
			fmt.Printf("警告: 下载文件 %s 失败: %v\n", node.FileName, err)
		} else {
			fmt.Printf("成功下载: %s\n", node.FileName)
		}
	}

	return nil
}

// downloadFile 简化的文件下载
func downloadFile(url, filepath string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
