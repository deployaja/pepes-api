package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	// Initialize database
	if err := InitDatabase(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Set Gin mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	r := gin.Default()

	// Configure CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins - you can restrict this to specific domains
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Routes for domains
	domains := r.Group("/domains")
	{
		domains.GET("", GetDomains)
		domains.POST("", CreateDomain)
		domains.GET(":id", GetDomain)
		domains.PUT(":id", UpdateDomain)
		domains.DELETE(":id", DeleteDomain)
	}

	// Routes for routes
	routes := r.Group("/routes")
	{
		routes.GET("", GetRoutes)
		routes.POST("", CreateRoute)
		routes.GET(":id", GetRoute)
		routes.PUT(":id", UpdateRoute)
		routes.DELETE(":id", DeleteRoute)
		routes.PUT(":id/plugins", UpdateRoutePlugin)
	}

	// Routes for plugins
	plugins := r.Group("/plugins")
	{
		plugins.GET("", GetPlugins)
		plugins.POST("", CreatePlugin)
		plugins.GET(":id", GetPlugin)
		plugins.PUT(":id", UpdatePlugin)
		plugins.DELETE(":id", DeletePlugin)
	}

	// Routes for plugin services
	pluginServices := r.Group("/plugin-services")
	{
		pluginServices.GET("", GetPluginServices)
		pluginServices.POST("", CreatePluginService)
		pluginServices.GET(":id", GetPluginService)
		pluginServices.PUT(":id", UpdatePluginService)
		pluginServices.DELETE(":id", DeletePluginService)
	}

	// Config endpoint (GET)
	r.GET("/config", GetConfig)

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("API server running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

func main() {
	Run()
}
