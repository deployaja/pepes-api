package main

import (
	"time"

	"gorm.io/gorm"
)

// Route represents a single route configuration in the database
type Route struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Path      string         `json:"path" gorm:"not null"`
	Upstream  string         `json:"upstream" gorm:"not null"`
	Plugin    string         `json:"plugin"`
	DomainID  uint           `json:"domain_id" gorm:"not null"`
	Domain    Domain         `json:"domain" gorm:"foreignKey:DomainID"`
	UsePathAsPrefix bool `json:"usePathAsPrefix"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Domain represents configuration for a single domain in the database
type Domain struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null;uniqueIndex"`
	UserId    string         `json:"user_id" gorm:"not null"`
	Routes    []Route        `json:"routes" gorm:"foreignKey:DomainID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Plugin represents a plugin configuration in the database
type Plugin struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	NamePlugin    string         `json:"name_plugin" gorm:"not null;uniqueIndex"`
	PluginSvcName string         `json:"plugin_svc_name" gorm:"not null"`
	Envs          string         `json:"envs" gorm:"type:text"`
	Desc          string         `json:"desc" gorm:"type:text"`
	UserId        string         `json:"user_id" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// CreateRouteRequest represents the request body for creating a route
type CreateRouteRequest struct {
	Path     string `json:"path" binding:"required,min=1,max=255" default:"/"`
	Upstream string `json:"upstream" binding:"required,min=1,max=500"`
	Plugin   string `json:"plugin" binding:"max=255"`
	DomainID uint   `json:"domain_id" binding:"required,min=1"`
	UsePathAsPrefix bool `json:"usePathAsPrefix" binding:"omitempty,boolean"`
}

// UpdateRouteRequest represents the request body for updating a route
type UpdateRouteRequest struct {
	Path     string  `json:"path" binding:"omitempty,min=1,max=255"`
	Upstream string  `json:"upstream" binding:"omitempty,min=1,max=500"`
	Plugin   *string `json:"plugin" binding:"omitempty,max=255"`
	DomainID uint    `json:"domain_id" binding:"omitempty,min=1"`
}

// UpdateRoutePluginRequest represents the request body for updating a route plugin
type UpdateRoutePluginRequest struct {
	Plugins *string `json:"plugins" binding:"omitempty,max=255"`
}

// CreateDomainRequest represents the request body for creating a domain
type CreateDomainRequest struct {
	Name    string `json:"name" binding:"required,min=1,max=255"`
	UserId  string `json:"user_id" binding:"required,min=1,max=255"`	
}

// UpdateDomainRequest represents the request body for updating a domain
type UpdateDomainRequest struct {
	Name string `json:"name" binding:"omitempty,min=1,max=255"`
}

// CreatePluginRequest represents the request body for creating a plugin
type CreatePluginRequest struct {
	NamePlugin    string `json:"name_plugin" binding:"required,min=1,max=255"`
	PluginSvcName string `json:"plugin_svc_name" binding:"required,min=1,max=255"`
	Envs          string `json:"envs" binding:"max=1000"`
	Desc          string `json:"desc" binding:"max=1000"`
	UserId        string `json:"user_id" binding:"required,min=1,max=255"`
}

// UpdatePluginRequest represents the request body for updating a plugin
type UpdatePluginRequest struct {
	NamePlugin    string `json:"name_plugin" binding:"omitempty,min=1,max=255"`
	PluginSvcName string `json:"plugin_svc_name" binding:"omitempty,min=1,max=255"`
	Envs          string `json:"envs" binding:"omitempty,max=1000"`
	Desc          string `json:"desc" binding:"omitempty,max=1000"`
}

// PluginService represents a plugin service configuration in the database
type PluginService struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name" gorm:"not null;uniqueIndex"`
	BaseConfig string         `json:"baseconfig" gorm:"type:json"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// CreatePluginServiceRequest represents the request body for creating a plugin service
type CreatePluginServiceRequest struct {
	Name       string `json:"name" binding:"required,min=1,max=255"`
	BaseConfig string `json:"baseconfig" binding:"max=5000"`
}

// UpdatePluginServiceRequest represents the request body for updating a plugin service
type UpdatePluginServiceRequest struct {
	Name       string `json:"name" binding:"omitempty,min=1,max=255"`
	BaseConfig string `json:"baseconfig" binding:"omitempty,max=5000"`
}
