package database

import (
	"fmt"
	"time"

	"github.com/SteaceP/coderage/internal/models"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() (*gorm.DB, error) {
	host := viper.GetString("database.host")
	if host == "" {
		return nil, fmt.Errorf("database host is not set")
	}

	port := viper.GetInt("database.port")
	if port == 0 {
		return nil, fmt.Errorf("database port is not set")
	}

	user := viper.GetString("database.user")
	if user == "" {
		return nil, fmt.Errorf("database user is not set")
	}

	password := viper.GetString("database.password")
	if password == "" {
		return nil, fmt.Errorf("database password is not set")
	}

	name := viper.GetString("database.name")
	if name == "" {
		return nil, fmt.Errorf("database name is not set")
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		name,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection pool: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil

}

func RunMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database pointer is nil, cannot run migrations")
	}

	// Auto migrate models
	err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
	)
	if err != nil {
		return fmt.Errorf("database migration failed: %v", err)
	}

	return nil
}
