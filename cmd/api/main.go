package main

import (
	"fmt"
	"log"
	"time"

	"nailly-back-end/internal/config"
	"nailly-back-end/internal/database"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/router"
	"nailly-back-end/internal/service"
)

func main() {
	setThailandTimezone()

	cfg := config.Load()
	if cfg.LineLoginChannelID == "" {
		log.Fatal("LINE_LOGIN_CHANNEL_ID is required for LINE LIFF customer login")
	}
	if cfg.CustomerJWTSecret == "" {
		log.Fatal("CUSTOMER_JWT_SECRET is required for customer JWT")
	}
	db := database.Connect(cfg.DSN)

	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Println("migrate users warning: ", err)
	} else {
		fmt.Println("Database User migrated!")
	}

	if err := db.AutoMigrate(&model.Admin{}); err != nil {
		log.Println("migrate admins warning: ", err)
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

	authService := service.NewAuthService(repository.NewAuthRepository(db), jwtManager)
	if err := authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminName, cfg.AdminPassword); err != nil {
		log.Println("seed admin warning: ", err)
	} else {
		fmt.Println("Database Admin migrated!")
	}

	if err := db.AutoMigrate(&model.Category{}); err != nil {
		log.Println("migrate categories warning: ", err)
	} else {
		fmt.Println("Database Category migrated!")
	}

	if err := db.AutoMigrate(&model.Service{}); err != nil {
		log.Println("migrate services warning: ", err)
	} else {
		fmt.Println("Database Service migrated!")
	}

	if err := db.AutoMigrate(&model.NailTechnician{}); err != nil {
		log.Println("migrate nail technicians warning: ", err)
	} else {
		fmt.Println("Database Nail Technician migrated!")
	}

	if err := database.EnsureCatalogImageSchema(db); err != nil {
		log.Println("prepare catalog image schema warning: ", err)
	}

	_ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_service_dbs_gorm_id ON service_dbs (id)").Error
	_ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_nail_technician_dbs_gorm_id ON nail_technician_dbs (id)").Error

	if db.Migrator().HasTable(&model.Booking{}) {
		_ = db.Exec("ALTER TABLE bookings ALTER COLUMN user_id DROP NOT NULL").Error
	}

	if err := database.EnsureDepositPaymentSchema(db); err != nil {
		log.Println("prepare deposit payment schema warning: ", err)
	}

	if err := db.AutoMigrate(&model.Booking{}); err != nil {
		log.Println("migrate bookings warning: ", err)
	} else {
		fmt.Println("Database Booking migrated!")
	}

	if err := db.AutoMigrate(&model.ShopSetting{}); err != nil {
		log.Println("migrate shop settings warning: ", err)
	} else {
		fmt.Println("Database Shop Setting migrated!")
	}

	r := router.New(db, cfg.AllowOrigin, jwtManager, customerJWTManager, cfg.LineLoginChannelID, lineNotificationService, slipUploader, imageUploader, qrCodeUploader)
	r.Run(":" + cfg.Port)
}

func setThailandTimezone() {
	time.Local = time.FixedZone("Asia/Bangkok", 7*60*60)
}
