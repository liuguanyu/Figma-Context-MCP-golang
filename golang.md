# Figma MCP Server - Golang重构分析与实现方案

## 目标
解决Node.js在处理大响应（~740KB）时的性能瓶颈，特别是`parseFigmaResponse`函数的效率问题，用Golang重新实现整个服务。

## 现有服务架构分析

### 1. 服务端点 (Endpoints)

#### 主要HTTP端点
- **`/health`** (GET) - 健康检查端点
  - 输出: `{"status":"healthy","timestamp":"...","server":"Figma MCP Server","version":"..."}`

- **`/mcp`** (POST) - StreamableHTTP MCP协议端点
  - 输入: MCP JSON-RPC协议消息
  - 输出: Server-Sent Events (SSE) 流
  - 支持会话管理 (MCP-Session-ID header)

#### 辅助端点
- **`/sse`** - SSE连接端点
- **`/messages`** - 消息端点

### 2. MCP协议工具 (Tools)

#### `get_figma_data`
- **描述**: 获取Figma文件的布局信息
- **输入参数**:
  ```json
  {
    "figmaApiKey": "string (required)" - Figma API认证密钥
    "fileKey": "string (required)" - Figma文件ID
    "nodeId": "string (optional)" - 特定节点ID
    "depth": "number (optional)" - 遍历深度
  }
  ```
- **输出**: 完整的Figma文件结构数据（经过简化处理）

#### `download_figma_images`  
- **描述**: 下载Figma文件中的SVG/PNG图像
- **输入参数**:
  ```json
  {
    "figmaApiKey": "string (required)",
    "fileKey": "string (required)", 
    "nodes": "array (required)" - 包含nodeId、fileName等的节点数组
    "localPath": "string (required)" - 本地存储路径
    "pngScale": "number (optional)" - PNG缩放比例
    "svgOptions": "object (optional)" - SVG导出选项
  }
  ```

### 3. 模块划分

#### 核心模块
1. **HTTP服务器** (`src/server.ts`)
   - Express.js HTTP服务器
   - 路由处理
   - SSE支持

2. **MCP协议处理** (`src/mcp.ts`)
   - MCP JSON-RPC协议实现
   - 工具注册和调度
   - 会话管理

3. **Figma服务** (`src/services/figma.ts`)
   - Figma API调用
   - 认证处理
   - 响应简化 ⚠️ **性能瓶颈位置**

4. **响应简化器** (`src/services/simplify-node-response.ts`)
   - 大JSON响应处理 ⚠️ **性能瓶颈位置**
   - 节点数据清理和转换

#### 工具模块  
1. **样式转换器** (`src/transformers/`)
   - 效果处理 (`effects.ts`)
   - 布局处理 (`layout.ts`) 
   - 样式处理 (`style.ts`)

2. **工具函数** (`src/utils/`)
   - 网络重试 (`fetch-with-retry.ts`)
   - 日志记录 (`logger.ts`)
   - 数据清理 (`sanitization.ts`)

### 4. 关键性能瓶颈分析

#### 主要问题: `parseFigmaResponse` 函数
位置: `src/services/figma.ts` 第约120行
```typescript
const simplifiedResponse = parseFigmaResponse(response);
```

#### 瓶颈原因:
1. **大JSON处理**: ~740KB的响应数据
2. **深度递归**: 复杂的节点树遍历
3. **内存分配**: JavaScript对象创建和GC压力
4. **单线程阻塞**: Node.js主线程阻塞

### 5. Figma API调用规范

#### 认证方式
- 使用Personal Access Token
- Header: `X-Figma-Token: {figmaApiKey}`

#### 主要API端点
1. **获取文件数据**: `GET https://api.figma.com/v1/files/{fileKey}`
2. **获取图像**: `GET https://api.figma.com/v1/images/{fileKey}`

## Golang实现方案

### 1. 架构设计

```
figma-mcp-golang/
├── main.go                 # 主入口
├── go.mod                  # 依赖管理
├── server/                 # HTTP服务器
│   ├── server.go          # 服务器主体
│   ├── handlers.go        # HTTP处理器
│   └── sse.go             # SSE实现
├── mcp/                    # MCP协议处理
│   ├── protocol.go        # 协议定义
│   ├── session.go         # 会话管理
│   └── tools.go           # 工具实现
├── figma/                  # Figma相关服务
│   ├── client.go          # API客户端
│   ├── parser.go          # 响应解析器 (性能优化重点)
│   └── downloader.go      # 图像下载器
├── types/                  # 类型定义
│   ├── figma.go           # Figma API类型
│   └── mcp.go             # MCP协议类型
└── utils/                  # 工具函数
    ├── logger.go          # 日志
    └── http.go            # HTTP工具
```

### 2. 性能优化策略

#### JSON处理优化
1. **流式解析**: 使用`encoding/json`的Decoder进行流式处理
2. **内存池**: 复用对象避免频繁分配
3. **并发处理**: 大数据结构的并行处理

#### 响应简化优化
1. **预分配切片**: 减少动态扩容
2. **直接映射**: 避免深拷贝
3. **懒加载**: 按需处理节点数据

#### 网络优化
1. **HTTP/2**: 支持多路复用
2. **连接池**: 复用TCP连接
3. **压缩**: 启用gzip压缩

### 3. 核心接口保持一致

#### HTTP接口
- 端口: 3333 (可配置)
- 路径: 完全兼容现有接口
- 响应格式: 保持SSE格式不变

#### MCP工具接口  
- 工具名称: `get_figma_data`, `download_figma_images`
- 参数结构: 完全兼容
- 响应格式: JSON-RPC 2.0标准

### 4. 关键实现要点

#### figmaApiKey处理
- ❌ 不从环境变量读取
- ✅ 直接从请求参数获取
- 每个工具调用都需要传递API key

#### 会话管理
- 支持MCP-Session-ID header
- 内存中维护会话状态
- 自动清理过期会话

#### 错误处理
- 统一错误响应格式
- Figma API错误传播
- 优雅降级处理

### 5. 部署和测试

#### 编译
```bash
cd figma-mcp-golang
go mod tidy
go build -o figma-mcp-server main.go
```

#### 运行
```bash
./figma-mcp-server -port 3333
```

#### 测试兼容性
- 现有的Java测试文件无需修改
- 所有测试用例应该通过
- 性能应有显著提升

## 预期性能提升

1. **响应时间**: 740KB响应处理时间从秒级降低到毫秒级
2. **内存使用**: Golang的更高效内存管理
3. **并发能力**: 原生协程支持更好的并发处理
4. **CPU使用率**: 更高效的JSON处理算法

## 实施步骤

1. ✅ 创建项目structure和分析文档
2. 🔄 实现基础HTTP服务器和健康检查
3. 🔄 实现MCP协议层
4. 🔄 实现Figma API客户端
5. 🔄 优化JSON解析器（重点）
6. 🔄 实现工具逻辑
7. 🔄 集成测试和性能验证
