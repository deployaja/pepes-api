package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// formatValidationError converts gin validation errors to human-readable messages
func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			field := getFieldDisplayName(fieldError.Field())
			tag := fieldError.Tag()

			switch tag {
			case "required":
				messages = append(messages, field+" is required")
			case "email":
				messages = append(messages, field+" must be a valid email address")
			case "min":
				messages = append(messages, field+" must be at least "+fieldError.Param()+" characters")
			case "max":
				messages = append(messages, field+" must be at most "+fieldError.Param()+" characters")
			case "url":
				messages = append(messages, field+" must be a valid URL")
			case "numeric":
				messages = append(messages, field+" must be a number")
			case "alpha":
				messages = append(messages, field+" must contain only letters")
			case "alphanum":
				messages = append(messages, field+" must contain only letters and numbers")
			default:
				messages = append(messages, field+" failed validation: "+tag)
			}
		}
		return strings.Join(messages, "; ")
	}

	// For non-validation errors, return the original error message
	return err.Error()
}

// formatDatabaseError converts database errors to human-readable messages
func formatDatabaseError(err error) string {
	if err == gorm.ErrRecordNotFound {
		return "Record not found"
	}

	// Check for unique constraint violations
	if strings.Contains(err.Error(), "duplicate key value") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "duplicate entry") {
		return "A record with this information already exists"
	}

	// Check for foreign key constraint violations
	if strings.Contains(err.Error(), "foreign key constraint") ||
		strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		return "Cannot delete this record as it is referenced by other records"
	}

	// Check for not null constraint violations
	if strings.Contains(err.Error(), "NOT NULL constraint failed") ||
		strings.Contains(err.Error(), "null value in column") {
		return "Required field cannot be empty"
	}

	// For other database errors, return a generic message
	return "Database operation failed. Please try again."
}

// getFieldDisplayName converts struct field names to user-friendly display names
func getFieldDisplayName(field string) string {
	fieldMap := map[string]string{
		"Path":          "Path",
		"Upstream":      "Upstream URL",
		"Plugin":        "Plugin",
		"DomainID":      "Domain ID",
		"Name":          "Name",
		"NamePlugin":    "Plugin Name",
		"PluginSvcName": "Plugin Service Name",
		"Envs":          "Environment Variables",
		"Desc":          "Description",
		"BaseConfig":    "Base Configuration",
	}

	if displayName, exists := fieldMap[field]; exists {
		return displayName
	}

	// If no mapping exists, convert camelCase to Title Case
	return strings.Title(strings.ToLower(field))
}

// GetRoutes returns all routes with optional filtering
func GetRoutes(c *gin.Context) {
	var routes []Route
	query := DB.Preload("Domain")

	// Filter by domain if provided
	if domainID := c.Query("domain_id"); domainID != "" {
		if id, err := strconv.ParseUint(domainID, 10, 32); err == nil {
			query = query.Where("domain_id = ?", id)
		}
	}

	// Filter by path if provided
	if path := c.Query("path"); path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}

	if err := query.Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  routes,
		"count": len(routes),
	})
}

// GetRoute returns a single route by ID
func GetRoute(c *gin.Context) {
	id := c.Param("id")
	var route Route

	if err := DB.Preload("Domain").First(&route, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": route})
}

// CreateRoute creates a new route
func CreateRoute(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	// Check if domain exists
	var domain Domain
	if err := DB.First(&domain, req.DomainID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	route := Route{
		Path:     req.Path,
		Upstream: req.Upstream,
		Plugin:   req.Plugin,
		DomainID: req.DomainID,
	}

	if err := DB.Create(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	// Load the domain relationship
	DB.Preload("Domain").First(&route, route.ID)

	c.JSON(http.StatusCreated, gin.H{"data": route})
}

func UpdateRoutePlugin(c *gin.Context) {
	id := c.Param("id")
	var route Route

	if err := DB.First(&route, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	var req UpdateRoutePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	if req.Plugins != nil {
		route.Plugin = *req.Plugins
	}

	if err := DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": route})
}

// UpdateRoute updates an existing route
func UpdateRoute(c *gin.Context) {
	id := c.Param("id")
	var route Route

	if err := DB.First(&route, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	var req UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	// Update fields if provided
	if req.Path != "" {
		route.Path = req.Path
	}
	if req.Upstream != "" {
		route.Upstream = req.Upstream
	}
	// For plugin field, update if provided (allows clearing by setting to empty string)
	if req.Plugin != nil {
		route.Plugin = *req.Plugin
	}
	if req.DomainID != 0 {
		// Check if new domain exists
		var domain Domain
		if err := DB.First(&domain, req.DomainID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
			return
		}
		route.DomainID = req.DomainID
	}

	if err := DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	// Load the domain relationship
	DB.Preload("Domain").First(&route, route.ID)

	c.JSON(http.StatusOK, gin.H{"data": route})
}

// DeleteRoute deletes a route
func DeleteRoute(c *gin.Context) {
	id := c.Param("id")
	var route Route

	if err := DB.First(&route, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	if err := DB.Delete(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted successfully"})
}

// GetDomains returns all domains
func GetDomains(c *gin.Context) {
	var domains []Domain
	query := DB.Preload("Routes")

	// Filter by name if provided
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  domains,
		"count": len(domains),
	})
}

// GetDomain returns a single domain by ID
func GetDomain(c *gin.Context) {
	id := c.Param("id")
	var domain Domain

	if err := DB.Preload("Routes").First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": domain})
}

// CreateDomain creates a new domain
func CreateDomain(c *gin.Context) {
	var req CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	domain := Domain{
		Name: req.Name,
		UserId: req.UserId,
	}

	if err := DB.Create(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": domain})
}

// UpdateDomain updates an existing domain
func UpdateDomain(c *gin.Context) {
	id := c.Param("id")
	var domain Domain

	if err := DB.First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	var req UpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	if req.Name != "" {
		domain.Name = req.Name
	}

	if err := DB.Save(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": domain})
}

// DeleteDomain deletes a domain
func DeleteDomain(c *gin.Context) {
	id := c.Param("id")
	var domain Domain

	if err := DB.Preload("Routes").First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	// Delete associated routes first
	if len(domain.Routes) > 0 {
		if err := DB.Delete(&domain.Routes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
			return
		}
	}

	if err := DB.Delete(&domain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain and associated routes deleted successfully"})
}

// GetPlugins returns all plugins with optional filtering
func GetPlugins(c *gin.Context) {
	var plugins []Plugin
	query := DB

	// Filter by name_plugin if provided
	if namePlugin := c.Query("name_plugin"); namePlugin != "" {
		query = query.Where("name_plugin LIKE ?", "%"+namePlugin+"%")
	}

	// Filter by plugin_svc_name if provided
	if pluginSvcName := c.Query("plugin_svc_name"); pluginSvcName != "" {
		query = query.Where("plugin_svc_name LIKE ?", "%"+pluginSvcName+"%")
	}

	if err := query.Find(&plugins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  plugins,
		"count": len(plugins),
	})
}

// GetPlugin returns a single plugin by ID
func GetPlugin(c *gin.Context) {
	id := c.Param("id")
	var plugin Plugin

	if err := DB.First(&plugin, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plugin})
}

// CreatePlugin creates a new plugin
func CreatePlugin(c *gin.Context) {
	var req CreatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	plugin := Plugin{
		NamePlugin:    req.NamePlugin,
		PluginSvcName: req.PluginSvcName,
		Envs:          req.Envs,
		Desc:          req.Desc,
		UserId:        req.UserId,
	}

	if err := DB.Create(&plugin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": plugin})
}

// UpdatePlugin updates an existing plugin
func UpdatePlugin(c *gin.Context) {
	id := c.Param("id")
	var plugin Plugin

	if err := DB.First(&plugin, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	var req UpdatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	// Update fields if provided
	if req.NamePlugin != "" {
		plugin.NamePlugin = req.NamePlugin
	}
	if req.PluginSvcName != "" {
		plugin.PluginSvcName = req.PluginSvcName
	}
	if req.Envs != "" {
		plugin.Envs = req.Envs
	}
	if req.Desc != "" {
		plugin.Desc = req.Desc
	}

	if err := DB.Save(&plugin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plugin})
}

// DeletePlugin deletes a plugin
func DeletePlugin(c *gin.Context) {
	id := c.Param("id")
	var plugin Plugin

	if err := DB.First(&plugin, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	if err := DB.Delete(&plugin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin deleted successfully"})
}

// GetConfig returns the configuration in the format expected by the original response.go
func GetConfig(c *gin.Context) {
	var domains []Domain
	if err := DB.Preload("Routes").Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	// Convert to the expected format
	config := ConfigResponse{
		Domains: make(map[string]DomainConfig),
	}

	for _, domain := range domains {
		domainConfig := DomainConfig{
			Routes: make([]RouteConfig, len(domain.Routes)),
		}
		var listPlugins []Plugin
		if err := DB.Find(&listPlugins).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
			return
		}

		for i, route := range domain.Routes {
			var filteredPlugins []Plugin
			for _, plugin := range listPlugins {
				if strings.Contains(route.Plugin, plugin.NamePlugin) {
					filteredPlugins = append(filteredPlugins, plugin)
				}
			}
			domainConfig.Routes[i] = RouteConfig{
				Path:        route.Path,
				Upstream:    route.Upstream,
				Plugin:      route.Plugin,
				PluginsData: filteredPlugins,
			}
		}

		config.Domains[domain.Name] = domainConfig
	}

	c.JSON(http.StatusOK, config)
}

// GetPluginServices returns all plugin services with optional filtering
func GetPluginServices(c *gin.Context) {
	var pluginServices []PluginService
	query := DB

	// Filter by name if provided
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Find(&pluginServices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  pluginServices,
		"count": len(pluginServices),
	})
}

// GetPluginService returns a single plugin service by ID
func GetPluginService(c *gin.Context) {
	id := c.Param("id")
	var pluginService PluginService

	if err := DB.First(&pluginService, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pluginService})
}

// CreatePluginService creates a new plugin service
func CreatePluginService(c *gin.Context) {
	var req CreatePluginServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	pluginService := PluginService{
		Name:       req.Name,
		BaseConfig: req.BaseConfig,
	}

	if err := DB.Create(&pluginService).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": pluginService})
}

// UpdatePluginService updates an existing plugin service
func UpdatePluginService(c *gin.Context) {
	id := c.Param("id")
	var pluginService PluginService

	if err := DB.First(&pluginService, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	var req UpdatePluginServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		pluginService.Name = req.Name
	}
	if req.BaseConfig != "" {
		pluginService.BaseConfig = req.BaseConfig
	}

	if err := DB.Save(&pluginService).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pluginService})
}

// DeletePluginService deletes a plugin service
func DeletePluginService(c *gin.Context) {
	id := c.Param("id")
	var pluginService PluginService

	if err := DB.First(&pluginService, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	if err := DB.Delete(&pluginService).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": formatDatabaseError(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin service deleted successfully"})
}
