package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"figma-mcp-server/server"
)

func main() {
	port := flag.Int("port", 3333, "服务器端口")
	flag.Parse()

	fmt.Printf("配置:\n")
	fmt.Printf("- 端口: %d\n", *port)
	fmt.Printf("- 认证方式: 从请求参数获取API Key\n")

	fmt.Printf("\n正在初始化 Figma MCP Server (HTTP 模式) 端口 %d...\n", *port)

	// 创建服务器
	srv := server.NewServer()

	// 启动服务器
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		fmt.Printf("[INFO] HTTP server listening on port %d\n", *port)
		fmt.Printf("[INFO] SSE endpoint available at http://localhost:%d/sse\n", *port)
		fmt.Printf("[INFO] Message endpoint available at http://localhost:%d/messages\n", *port)
		fmt.Printf("[INFO] StreamableHTTP endpoint available at http://localhost:%d/mcp\n", *port)

		if err := http.ListenAndServe(addr, srv); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n[INFO] 服务器正在关闭...")
}
