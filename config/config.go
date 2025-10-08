package config

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env"
	pmodel "github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres/model"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Config struct {
	Port        int    `env:"PORT" envDefault:"8080"`
	JwtSecret   string `env:"JWT_SECRET"`
	DatabaseURL string `env:"DATABASE_URL"`
	Schema      string `env:"SCHEMA" envDefault:"public"`
	SkipMigrate bool   `env:"SKIP_MIGRATE" envDefault:"false"` // Add option to skip migration
}

var (
	configuration Config
)

func LoadEnv() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or failed to load .env file: %v", err)
	}

	if err := env.Parse(&configuration); err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}

	return configuration
}

func GetConfig() Config {
	return configuration
}

func InitDB(dbUrl, schemaName string) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbUrl,
		PreferSimpleProtocol: true, // Use simple protocol to avoid prepared statements
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              false,                                // Disable prepared statements
		Logger:                                   logger.Default.LogMode(logger.Error), // Only log errors
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			TablePrefix:   fmt.Sprintf("%s.", schemaName),
		},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database instance: %v", err)
	}

	// Very aggressive connection pool settings to minimize reuse
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetConnMaxLifetime(30 * time.Second)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Only migrate if not explicitly skipped
	cfg := GetConfig()
	if !cfg.SkipMigrate {
		log.Println("Running database migration...")
		migrate(db, schemaName)
		log.Println("Migration completed successfully!")
	} else {
		log.Println("Skipping database migration (SKIP_MIGRATE=true)")
	}

	return db
}

func migrate(db *gorm.DB, schemaName string) {
	err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schemaName).Error
	if err != nil {
		log.Fatal("failed to create schema:", err)
	}

	models := []interface{}{
		&pmodel.User{},
		&pmodel.Client{},
		&pmodel.Invoice{},
		&pmodel.InvoiceItem{},
	}

	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			log.Printf("Creating table for %T...", model)
		} else {
			log.Printf("Table for %T already exists, checking for updates...", model)
		}
	}

	if err := db.AutoMigrate(models...); err != nil {
		log.Printf("Migration warning: %v", err)
	}
}
