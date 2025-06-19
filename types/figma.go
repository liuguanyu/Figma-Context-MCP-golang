package types

// Figma数据结构定义，对应NodeJS版本的SimplifiedDesign和相关类型

type SimplifiedDesign struct {
	Name          string                 `json:"name"`
	LastModified  string                 `json:"lastModified"`
	ThumbnailUrl  string                 `json:"thumbnailUrl"`
	Nodes         []SimplifiedNode       `json:"nodes"`
	Components    map[string]interface{} `json:"components"`
	ComponentSets map[string]interface{} `json:"componentSets"`
	GlobalVars    GlobalVars             `json:"globalVars"`
}

type SimplifiedNode struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Type                string              `json:"type"`
	Text                string              `json:"text,omitempty"`
	TextStyle           string              `json:"textStyle,omitempty"`
	Fills               string              `json:"fills,omitempty"`
	Styles              string              `json:"styles,omitempty"`
	Strokes             string              `json:"strokes,omitempty"`
	Effects             string              `json:"effects,omitempty"`
	Opacity             *float64            `json:"opacity,omitempty"`
	BorderRadius        string              `json:"borderRadius,omitempty"`
	Layout              string              `json:"layout,omitempty"`
	ComponentId         string              `json:"componentId,omitempty"`
	ComponentProperties []ComponentProperty `json:"componentProperties,omitempty"`
	Children            []SimplifiedNode    `json:"children,omitempty"`
}

type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ComponentProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type GlobalVars struct {
	Styles map[string]interface{} `json:"styles"`
}

// Layout structures to match NodeJS implementation
type SimplifiedLayout struct {
	Mode                     string                 `json:"mode"`
	JustifyContent           string                 `json:"justifyContent,omitempty"`
	AlignItems               string                 `json:"alignItems,omitempty"`
	AlignSelf                string                 `json:"alignSelf,omitempty"`
	Wrap                     *bool                  `json:"wrap,omitempty"`
	Gap                      string                 `json:"gap,omitempty"`
	LocationRelativeToParent map[string]interface{} `json:"locationRelativeToParent,omitempty"`
	Dimensions               map[string]interface{} `json:"dimensions,omitempty"`
	Padding                  string                 `json:"padding,omitempty"`
	Sizing                   map[string]interface{} `json:"sizing,omitempty"`
	OverflowScroll           []string               `json:"overflowScroll,omitempty"`
	Position                 string                 `json:"position,omitempty"`
}

type FigmaGetFileResult struct {
	Metadata   SimplifiedDesignMetadata `json:"metadata"`
	Nodes      []SimplifiedNode         `json:"nodes"`
	GlobalVars GlobalVars               `json:"globalVars"`
}

type SimplifiedDesignMetadata struct {
	Name         string `json:"name"`
	LastModified string `json:"lastModified"`
	ThumbnailUrl string `json:"thumbnailUrl"`
}

// Figma API原始响应结构
type FigmaAPIResponse struct {
	Name          string                 `json:"name"`
	LastModified  string                 `json:"lastModified"`
	ThumbnailUrl  string                 `json:"thumbnailUrl"`
	Document      FigmaNode              `json:"document,omitempty"`
	Components    map[string]interface{} `json:"components,omitempty"`
	ComponentSets map[string]interface{} `json:"componentSets,omitempty"`
}

type FigmaAPINodeResponse struct {
	Name          string                      `json:"name"`
	LastModified  string                      `json:"lastModified"`
	ThumbnailUrl  string                      `json:"thumbnailUrl"`
	Nodes         map[string]FigmaNodeWrapper `json:"nodes"`
	Components    map[string]interface{}      `json:"components,omitempty"`
	ComponentSets map[string]interface{}      `json:"componentSets,omitempty"`
}

type FigmaNodeWrapper struct {
	Document      FigmaNode              `json:"document"`
	Components    map[string]interface{} `json:"components,omitempty"`
	ComponentSets map[string]interface{} `json:"componentSets,omitempty"`
}

type FigmaNode struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"`
	Visible              *bool                  `json:"visible,omitempty"`
	AbsoluteBoundingBox  *BoundingBox           `json:"absoluteBoundingBox,omitempty"`
	Characters           string                 `json:"characters,omitempty"`
	Style                map[string]interface{} `json:"style,omitempty"`
	Fills                []interface{}          `json:"fills,omitempty"`
	Strokes              []interface{}          `json:"strokes,omitempty"`
	StrokeWeight         float64                `json:"strokeWeight,omitempty"`
	StrokeAlign          string                 `json:"strokeAlign,omitempty"`
	Effects              []interface{}          `json:"effects,omitempty"`
	Opacity              *float64               `json:"opacity,omitempty"`
	CornerRadius         *float64               `json:"cornerRadius,omitempty"`
	RectangleCornerRadii []float64              `json:"rectangleCornerRadii,omitempty"`
	ComponentId          string                 `json:"componentId,omitempty"`
	ComponentProperties  map[string]interface{} `json:"componentProperties,omitempty"`
	Children             []FigmaNode            `json:"children,omitempty"`

	// Layout properties for frame nodes
	LayoutMode             string   `json:"layoutMode,omitempty"`
	PrimaryAxisAlignItems  string   `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems  string   `json:"counterAxisAlignItems,omitempty"`
	LayoutAlign            string   `json:"layoutAlign,omitempty"`
	LayoutWrap             string   `json:"layoutWrap,omitempty"`
	ItemSpacing            *float64 `json:"itemSpacing,omitempty"`
	PaddingTop             *float64 `json:"paddingTop,omitempty"`
	PaddingRight           *float64 `json:"paddingRight,omitempty"`
	PaddingBottom          *float64 `json:"paddingBottom,omitempty"`
	PaddingLeft            *float64 `json:"paddingLeft,omitempty"`
	LayoutSizingHorizontal string   `json:"layoutSizingHorizontal,omitempty"`
	LayoutSizingVertical   string   `json:"layoutSizingVertical,omitempty"`
	LayoutPositioning      string   `json:"layoutPositioning,omitempty"`
	LayoutGrow             *float64 `json:"layoutGrow,omitempty"`
	OverflowDirection      []string `json:"overflowDirection,omitempty"`
	PreserveRatio          bool     `json:"preserveRatio,omitempty"`
}

// 图像下载相关的类型定义
type ImageNode struct {
	NodeId   string `json:"nodeId"`
	ImageRef string `json:"imageRef,omitempty"`
	FileName string `json:"fileName"`
}

type SVGOptions struct {
	OutlineText    bool `json:"outlineText"`
	IncludeId      bool `json:"includeId"`
	SimplifyStroke bool `json:"simplifyStroke"`
}

// Figma Images API响应
type FigmaImagesResponse struct {
	Images map[string]string `json:"images"`
	Error  string            `json:"error,omitempty"`
}

// Figma Image Fills API响应
type FigmaImageFillsResponse struct {
	Meta struct {
		Images map[string]string `json:"images"`
	} `json:"meta"`
	Error string `json:"error,omitempty"`
}
