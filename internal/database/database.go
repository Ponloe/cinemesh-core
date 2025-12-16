package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Ponloe/cinemesh-core/internal/users"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s password=%s",
		host, user, dbname, port, sslmode, os.Getenv("DB_PASSWORD"))

	log.Printf("connecting to database host=%s db=%s user=%s port=%s sslmode=%s", host, dbname, user, port, sslmode)

	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("db.DB(): %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	DB = db
	log.Println("database connection established")
	return nil
}

func Migrate() error {
	if DB == nil {
		return errors.New("database not connected")
	}
	log.Println("running AutoMigrate for users.User")
	if err := DB.AutoMigrate(&users.User{}); err != nil {
		return fmt.Errorf("auto migrate users: %w", err)
	}
	log.Println("migrations complete")
	return nil
}

func Seed() error {
	if os.Getenv("SEED_DATA") != "true" {
		return nil
	}
	if DB == nil {
		return errors.New("database not connected")
	}

	if !DB.Migrator().HasTable(&users.User{}) {
		log.Println("users table missing, running AutoMigrate before seeding")
		if err := DB.AutoMigrate(&users.User{}); err != nil {
			return fmt.Errorf("auto migrate users before seed: %w", err)
		}
	}

	var u users.User
	err := DB.First(&u, "email = ?", "admin@cinemesh.com").Error
	if err == nil {
		log.Println("admin user already present")
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		admin := users.User{
			Username:     "admin",
			Email:        "admin@cinemesh.com",
			PasswordHash: "",
			Role:         "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			return fmt.Errorf("create admin: %w", err)
		}
		log.Println("admin user created")
		return nil
	}
	return fmt.Errorf("seed query: %w", err)
}
