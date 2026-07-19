package database

import (
	"errors"
	"finalproject/config"
	"finalproject/models"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func StartDB() error {
	cfg := config.Load()
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if cfg.GinMode != "release" {
		conn = conn.Debug()
	}

	if err := conn.AutoMigrate(models.User{}, models.Photo{}, models.Comment{}, models.SocialMedia{}, models.PushSubscription{}); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	if err := seedBootstrapUsers(conn, cfg); err != nil {
		return fmt.Errorf("seed bootstrap users: %w", err)
	}

	db = conn
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func SetDB(database *gorm.DB) {
	db = database
}

func seedBootstrapUsers(conn *gorm.DB, cfg config.Config) error {
	if err := seedBootstrapUser(
		conn,
		"admin",
		cfg.BootstrapAdminEmail,
		cfg.BootstrapAdminUsername,
		cfg.BootstrapAdminPassword,
		cfg.BootstrapAdminAge,
		models.RoleAdmin,
		true,
	); err != nil {
		return err
	}

	return seedBootstrapUser(
		conn,
		"user",
		cfg.BootstrapUserEmail,
		cfg.BootstrapUserUsername,
		cfg.BootstrapUserPassword,
		cfg.BootstrapUserAge,
		models.RoleUser,
		false,
	)
}

func seedBootstrapUser(
	conn *gorm.DB,
	label string,
	email string,
	username string,
	password string,
	age int,
	role string,
	ensureRole bool,
) error {
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if email == "" && username == "" && password == "" {
		return nil
	}

	if email == "" || username == "" || password == "" {
		return fmt.Errorf("bootstrap %s requires email, username, and password", label)
	}

	var existing models.User
	err := conn.Where("email = ? OR username = ?", email, username).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return conn.Create(&models.User{
			Username: username,
			Email:    email,
			Password: password,
			Age:      age,
			Role:     role,
			Status:   models.UserStatusActive,
		}).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{}
	if ensureRole && existing.Role != role {
		updates["role"] = role
	}
	if existing.Status != models.UserStatusActive {
		updates["status"] = models.UserStatusActive
		updates["banned_at"] = nil
		updates["ban_reason"] = ""
	}

	if len(updates) == 0 {
		return nil
	}

	return conn.Model(&existing).Updates(updates).Error
}
