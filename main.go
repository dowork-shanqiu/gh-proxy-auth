package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dowork-shanqiu/gh-proxy-auth/internal/config"
	"github.com/dowork-shanqiu/gh-proxy-auth/internal/database"
	"github.com/dowork-shanqiu/gh-proxy-auth/internal/frontend"
	"github.com/dowork-shanqiu/gh-proxy-auth/internal/router"
	"github.com/dowork-shanqiu/gh-proxy-auth/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// Load config
	if err := config.Load(*configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// Initialize WebAuthn
	if err := service.InitWebAuthn(); err != nil {
		log.Fatalf("初始化 WebAuthn 失败: %v", err)
	}

	// Initialize embedded frontend
	if err := frontend.Init(); err != nil {
		log.Printf("警告: 加载前端资源失败: %v", err)
	}

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	// Setup routes
	router.Setup(engine)

	// Start server
	addr := fmt.Sprintf("%s:%d", config.C.Server.Host, config.C.Server.Port)
	log.Printf("服务启动在 %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
