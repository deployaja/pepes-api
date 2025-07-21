package main

// RouteConfig represents a single route configuration
type RouteConfig struct {
	Path     string `json:"path"`
	Upstream string `json:"upstream"`
	Plugin   string `json:"plugin"`
	PluginsData []Plugin `json:"plugins_data"`
}

// DomainConfig represents configuration for a single domain
type DomainConfig struct {
	Routes []RouteConfig `json:"routes"`
}

// ConfigResponse represents the complete configuration response
type ConfigResponse struct {
	Domains map[string]DomainConfig `json:"domains"`
}
