package users

import (
	"UrlShortener/internal/config"
	"UrlShortener/internal/models"
	"UrlShortener/internal/mymiddleware"
	"UrlShortener/internal/routing"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var db *gorm.DB

func InitDB() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	err = godotenv.Load("config/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	pass := os.Getenv("POSTGRES_PASSWORD")
	if pass == "" {
		log.Fatalf("POSTGRES_PASSWORD not set in .env file")
	}
	// Печать значений конфигурации для отладки
	fmt.Printf("DB Config: Host=%s, Port=%d, Username=%s, Database=%s\n",
		cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Username, cfg.Storage.Database)

	// Формирование строки подключения (DSN)
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%d sslmode=disable password=%s",
		cfg.Storage.Host, cfg.Storage.Username, cfg.Storage.Database, cfg.Storage.Port, pass)

	// Подключение к базе данных с использованием GORM
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция для создания таблиц
	err = db.AutoMigrate(&models.User{}, &models.UserProfile{}, &models.Verification{}, &models.LinkDB{}, &models.IPLINK{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	mymiddleware.InitializeDB(db)
	// Установка базы данных в пакете routing
	routing.SetDB(db)
}
