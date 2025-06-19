# Figma MCP Golang 服务器

这是一个使用 Go 语言实现的 Figma Model Context Protocol (MCP) 服务器，提供 HTTP/SSE 接口来访问 Figma API 功能。

## 项目简介

该服务器实现了 MCP 协议，可以作为远程服务运行，允许多个客户端通过 HTTP 接口访问 Figma API 功能。与传统的 MCP 服务器不同，本服务器采用了**按请求传递 API 密钥**的设计，而不是在启动时预配置，这使得它更适合多用户环境。

## 致谢

本项目受到了 [Figma-Context-MCP](https://github.com/GLips/Figma-Context-MCP) 项目的启发。原项目是一个优秀的 TypeScript/Node.js 实现的 Figma MCP 服务器，为 Figma 设计文件提供了强大的 AI 上下文支持。本 Golang 版本在保持核心功能的基础上，重新设计了架构以支持多用户并发访问，并采用 HTTP/SSE 接口提供更好的可扩展性。

感谢原项目作者的创新和贡献，为 Figma 生态系统带来了 MCP 协议的支持。

## 主要特性

- **多用户友好**: 每个请求都可以携带不同的 Figma API Key，支持多用户并发使用
- **HTTP/SSE 接口**: 提供标准的 HTTP REST API 和 Server-Sent Events 支持
- **MCP 协议兼容**: 完全兼容 Model Context Protocol 规范
- **图像下载支持**: 支持下载 Figma 文件中的 SVG 和 PNG 图像
- **会话管理**: 支持会话管理和状态跟踪

## 支持的工具

### 1. get_figma_data
获取 Figma 文件的布局信息和节点数据。

**参数:**
- `figmaApiKey` (必需): Figma API 认证密钥
- `fileKey` (必需): Figma 文件 ID
- `nodeId` (可选): 特定节点 ID
- `depth` (可选): 遍历深度

### 2. download_figma_images
下载 Figma 文件中的 SVG 和 PNG 图像。

**参数:**
- `figmaApiKey` (必需): Figma API 认证密钥
- `fileKey` (必需): Figma 文件 ID
- `nodes` (必需): 包含 nodeId、fileName 等的节点数组
- `localPath` (必需): 本地存储路径
- `pngScale` (可选): PNG 缩放比例，默认为 1.0
- `svgOptions` (可选): SVG 导出选项

## 安装和编译

### 环境要求
- Go 1.21 或更高版本
- 网络连接以访问 Figma API

### 编译步骤

1. 克隆或下载源代码到本地
2. 进入项目目录：
   ```bash
   cd figma-mcp-golang
   ```

3. 下载依赖：
   ```bash
   go mod download
   ```

4. 编译可执行文件：
   ```bash
   go build -o figma-mcp-server .
   ```

## 运行服务器

### 基本运行
```bash
./figma-mcp-server
```

### 指定端口运行
```bash
./figma-mcp-server -port 8080
```

### 运行参数
- `-port`: 指定服务器运行端口，默认为 3333

## API 端点

服务器运行后，将提供以下端点：

- **健康检查**: `GET /health`
- **MCP 主端点**: `POST /mcp` (支持 StreamableHTTP)
- **SSE 连接**: `GET /sse`
- **消息处理**: `POST /messages`

## 使用示例

### 获取 Figma 文件数据

```bash
curl -X POST http://localhost:3333/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_figma_data",
      "arguments": {
        "figmaApiKey": "your-figma-api-key",
        "fileKey": "your-file-key",
        "nodeId": "optional-node-id",
        "depth": 2
      }
    }
  }'
```

### 下载 Figma 图像

```bash
curl -X POST http://localhost:3333/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "download_figma_images",
      "arguments": {
        "figmaApiKey": "your-figma-api-key",
        "fileKey": "your-file-key",
        "nodes": [
          {
            "nodeId": "1234:5678",
            "fileName": "icon.svg"
          }
        ],
        "localPath": "/path/to/save/images",
        "pngScale": 2.0
      }
    }
  }'
```

## 项目结构

```
figma-mcp-golang/
├── main.go              # 主程序入口
├── go.mod              # Go 模块定义
├── go.sum              # 依赖校验和
├── figma/              # Figma API 客户端
│   └── client.go
├── mcp/                # MCP 协议实现
│   └── tools.go
├── server/             # HTTP 服务器实现
│   ├── handlers.go     # 请求处理器
│   ├── server.go       # 服务器核心
│   └── session.go      # 会话管理
└── types/              # 类型定义
    ├── figma.go        # Figma 相关类型
    └── mcp.go          # MCP 相关类型
```

## 多用户设计的优势

传统的 MCP 服务器通常在启动时配置 API 密钥，这意味着所有用户必须共享同一个 API 密钥。而本服务器采用了**请求级别的 API 密钥传递**设计：

1. **隔离性**: 每个用户可以使用自己的 Figma API 密钥
2. **安全性**: API 密钥不需要存储在服务器端
3. **灵活性**: 支持不同权限级别的用户访问不同的 Figma 文件
4. **可扩展性**: 易于部署为共享服务，支持多租户使用

## MCP 协议支持

本服务器完全实现了 MCP (Model Context Protocol) 规范，支持：

- **协议版本**: 2024-11-05
- **工具调用**: 支持动态工具发现和调用
- **会话管理**: 支持有状态的会话连接
- **错误处理**: 标准的 JSON-RPC 错误响应格式

## 配置和部署

### 开发环境
直接运行编译后的可执行文件即可用于开发和测试。

### 生产环境
建议配合反向代理（如 Nginx）使用，并设置适当的超时和并发限制。

### Docker 部署
可以创建 Dockerfile 进行容器化部署：

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o figma-mcp-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/figma-mcp-server .
EXPOSE 3333
CMD ["./figma-mcp-server"]
```

## 故障排除

### 常见问题

1. **端口占用**: 如果默认端口 3333 被占用，使用 `-port` 参数指定其他端口
2. **网络连接**: 确保服务器能够访问 Figma API (api.figma.com)
3. **API 密钥错误**: 检查 Figma API 密钥是否有效且有相应文件的访问权限

### 日志输出
服务器会输出详细的日志信息，可以通过日志排查问题：
- 连接信息
- 请求处理状态
- 错误信息

## 版本信息

- **服务器版本**: 0.4.2
- **Go 版本要求**: 1.21+
- **MCP 协议版本**: 2024-11-05

## 依赖项

- `github.com/gorilla/mux`: HTTP 路由器
- `github.com/rs/cors`: CORS 中间件
- `gopkg.in/yaml.v2`: YAML 配置解析

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

请查看项目根目录的许可证文件了解使用条款。
