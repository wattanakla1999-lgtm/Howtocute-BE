package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithError(dsn string) (*gorm.DB, error) {
	thailandLocation := time.FixedZone("Asia/Bangkok", 7*60*60)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(thailandLocation)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Database connected!")
	return db, nil
}

func Connect(dsn string) *gorm.DB {
	db, err := ConnectWithError(dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

