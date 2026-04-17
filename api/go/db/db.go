package db

import (
	"fmt"
	"image-to-palette/model"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func InitDatabase() error {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	config := getDatabaseConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

func getDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		Database: getEnv("DB_NAME", "image_to_palette"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrations() error {
	if dbError := DB.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE schemaname = 'public' AND indexname = 'idx_users_name'
			) THEN
				DROP INDEX IF EXISTS idx_users_name;
			END IF;
		END $$;
	`).Error; dbError != nil {
		return dbError
	}

	if err := DB.AutoMigrate(&model.Palette{}, &model.Theme{}, &model.User{}, &model.UserPreferences{}); err != nil {
		return err
	}

	if !DB.Migrator().HasColumn(&model.Palette{}, "is_shared") {
		if err := DB.Migrator().AddColumn(&model.Palette{}, "IsShared"); err != nil {
			return err
		}
	}
	if !DB.Migrator().HasColumn(&model.Palette{}, "shared_at") {
		if err := DB.Migrator().AddColumn(&model.Palette{}, "SharedAt"); err != nil {
			return err
		}
	}

	if !DB.Migrator().HasColumn(&model.Theme{}, "is_shared") {
		if err := DB.Migrator().AddColumn(&model.Theme{}, "IsShared"); err != nil {
			return err
		}
	}
	if !DB.Migrator().HasColumn(&model.Theme{}, "shared_at") {
		if err := DB.Migrator().AddColumn(&model.Theme{}, "SharedAt"); err != nil {
			return err
		}
	}

	return nil
}

func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
