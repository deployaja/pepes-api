package main

type PluginData struct {
	ID            int     `json:"id"`
	NamePlugin    string  `json:"name_plugin"`
	PluginSvcName string  `json:"plugin_svc_name"`
	Envs          string  `json:"envs"`
	Desc          string  `json:"desc"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	DeletedAt     *string `json:"deleted_at"`
}

// RouteConfig represents a single route configuration
type RouteConfig struct {
	Path            string       `json:"path"`
	Upstream        string       `json:"upstream"`
	Plugin          string       `json:"plugin"`
	UsePathAsPrefix bool         `json:"usePathAsPrefix"`
	PluginsData     []PluginData `json:"plugins_data"`
}

// DomainConfig represents configuration for a single domain
type DomainConfig struct {
	Routes []RouteConfig `json:"routes"`
}

// ConfigResponse represents the complete configuration response
type ConfigResponse struct {
	Domains map[string]DomainConfig `json:"domains"`
}
