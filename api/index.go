package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"nailly-back-end/internal/config"
	"nailly-back-end/internal/database"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/router"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
)

var (
	app     *gin.Engine
	initErr error
	once    sync.Once
)

func initApp() {
	defer func() {
		if r := recover(); r != nil {
			initErr = fmt.Errorf("panic during init: %v", r)
			log.Println("CRITICAL ERROR during Vercel init:", initErr)
		}
	}()

	time.Local = time.FixedZone("Asia/Bangkok", 7*60*60)
	gin.SetMode(gin.ReleaseMode)

	cfg := config.Load()

	db, err := database.ConnectWithError(cfg.DSN)
	if err != nil {
		initErr = fmt.Errorf("database connection failed: %w", err)
		log.Println("CRITICAL ERROR:", initErr)
		return
	}

	jwtManager := service.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)
	customerJWTManager := service.NewCustomerJWTManager(cfg.CustomerJWTSecret, cfg.CustomerJWTTTL)
	lineNotificationService := service.NewLineNotificationService(
		service.NewHTTPLinePushClient(nil),
		cfg.LineMessagingChannelAccessToken,
		cfg.LineShopOwnerUserID,
		cfg.LineBookingDetailsURL,
	)

	var slipUploader service.SlipUploader
	if storageClient := service.NewSupabaseStorageClient(
		cfg.SupabaseURL,
		cfg.SupabaseServiceRoleKey,
		cfg.SupabaseStorageBucket,
	); storageClient != nil {
		slipUploader = storageClient
	}

	var imageUploader service.ImageUploader
	if storageClient := service.NewSupabaseStorageClient(
		cfg.SupabaseURL,
		cfg.SupabaseServiceRoleKey,
		cfg.SupabaseProfileImageBucket,
	); storageClient != nil {
		imageUploader = storageClient
	}

	var qrCodeUploader service.ImageUploader
	if storageClient := service.NewSupabaseStorageClient(
		cfg.SupabaseURL,
		cfg.SupabaseServiceRoleKey,
		"QRCodeOwner",
	); storageClient != nil {
		qrCodeUploader = storageClient
	}

	// Only run migrations locally, skip DDL lock operations on Vercel PgBouncer pooler
	if os.Getenv("VERCEL") == "" && os.Getenv("VERCEL_ENV") == "" {
		_ = db.AutoMigrate(&model.User{})
		_ = db.AutoMigrate(&model.Admin{})
		authService := service.NewAuthService(repository.NewAuthRepository(db), jwtManager)
		_ = authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminName, cfg.AdminPassword)

		_ = db.AutoMigrate(&model.Service{})
		_ = db.AutoMigrate(&model.NailTechnician{})
		_ = database.EnsureCatalogImageSchema(db)

		_ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_service_dbs_gorm_id ON service_dbs (id)").Error
		_ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_nail_technician_dbs_gorm_id ON nail_technician_dbs (id)").Error
		if db.Migrator().HasTable(&model.Booking{}) {
			_ = db.Exec("ALTER TABLE bookings ALTER COLUMN user_id DROP NOT NULL").Error
		}
		_ = database.EnsureDepositPaymentSchema(db)
		_ = db.AutoMigrate(&model.Booking{})
		_ = db.AutoMigrate(&model.ShopSetting{})
	}

	app = router.New(db, cfg.AllowOrigin, jwtManager, customerJWTManager, cfg.LineLoginChannelID, lineNotificationService, slipUploader, imageUploader, qrCodeUploader)
	fmt.Println("Vercel Serverless Function initialized successfully")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initApp)
	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"error","message":%q}`, initErr.Error())
		return
	}

	// Restore original request path for Gin routing when Vercel rewrites to /api/main
	if origURI := r.Header.Get("x-forwarded-uri"); origURI != "" {
		r.URL.Path = origURI
	}

	app.ServeHTTP(w, r)
}
