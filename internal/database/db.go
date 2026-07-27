package database

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithError(dsn string) (*gorm.DB, error) {
	thailandLocation := time.FixedZone("Asia/Bangkok", 7*60*60)

	isVercel := os.Getenv("VERCEL") != "" || os.Getenv("VERCEL_ENV") != ""
	if isVercel && (os.Getenv("DATABASE_DSN") == "" || strings.Contains(dsn, "host=localhost") || strings.Contains(dsn, "127.0.0.1")) {
		return nil, fmt.Errorf("DATABASE_DSN is missing in Vercel environment variables. Please add DATABASE_DSN in Vercel Dashboard -> Settings -> Environment Variables")
	}

	if dsn != "" && !strings.Contains(dsn, "connect_timeout") {
		if strings.Contains(dsn, "?") {
			dsn += "&connect_timeout=5"
		} else if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			dsn += "?connect_timeout=5"
		} else {
			dsn += " connect_timeout=5"
		}
	}

	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database DSN: %w", err)
	}

	// Force IPv4 tcp4 network dialing for Vercel / serverless compatibility
	connConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var d net.Dialer
		d.Timeout = 5 * time.Second
		return d.DialContext(ctx, "tcp4", addr)
	}

	dbKey := stdlib.RegisterConnConfig(connConfig)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DriverName:           "pgx",
		DSN:                  dbKey,
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
