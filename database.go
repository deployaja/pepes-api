package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() error {
	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "gate_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Open database connection
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run auto migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Database connected and migrations completed successfully")
	return nil
}

// runMigrations runs the database migrations
func runMigrations() error {
	SeedPluginServices()
	// Auto migrate the models
	err := DB.AutoMigrate(&Domain{}, &Route{}, &Plugin{}, &PluginService{})
	if err != nil {
		return err
	}

	// Create indexes for better performance
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_routes_domain_id ON routes(domain_id)").Error
	if err != nil {
		return err
	}

	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name)").Error
	if err != nil {
		return err
	}

	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_plugins_name_plugin ON plugins(name_plugin)").Error
	if err != nil {
		return err
	}

	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_plugin_services_name ON plugin_services(name)").Error
	if err != nil {
		return err
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SeedPluginServices() error {
	pluginServices := []PluginService{
		{
			Name:       "auth",
			BaseConfig: `{"auth_user":"admin","auth_pass":"password"}`,
		},
		{
			Name:       "cors",
			BaseConfig: `{"cors_origin":"*","cors_methods":"GET,POST","cors_headers":"Content-Type,Authorization"}`,
		},
		{
			Name:       "ratelimit",
			BaseConfig: `{"rate_limit":60,"rate_window":60}`,
		},
		{
			Name:       "ipwhitelist",
			BaseConfig: `{"whitelist_ips":"127.0.0.1,192.168.1.0/24"}`,
		},
		{
			Name:       "jwt",
			BaseConfig: `{"jwt_secret":"mysecret"}`,
		},
		{
			Name:       "logging",
			BaseConfig: `{"log_level":"info"}`,
		},	
	}

	for _, svc := range pluginServices {
		var existing PluginService
		// Check if the plugin service already exists
		if err := DB.Where("name = ?", svc.Name).First(&existing).Error; err == nil {
			continue // already exists, skip
		}
		if err := DB.Create(&svc).Error; err != nil {
			log.Printf("Failed to seed plugin service %s: %v", svc.Name, err)
			return err
		}
		log.Printf("Seeded plugin service: %s", svc.Name)
	}
	return nil
}