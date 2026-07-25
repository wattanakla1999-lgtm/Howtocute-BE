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
		log.Fatal("migrate users: ", err)
	}
	fmt.Println("Database User migrated!")
	if err := db.AutoMigrate(&model.Admin{}); err != nil {
		log.Fatal("migrate admins: ", err)
	}
	jwtManager := service.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)
	customerJWTManager := service.NewCustomerJWTManager(cfg.CustomerJWTSecret, cfg.CustomerJWTTTL)
	authService := service.NewAuthService(repository.NewAuthRepository(db), jwtManager)
	if err := authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminName, cfg.AdminPassword); err != nil {
		log.Fatal("seed configured admin: ", err)
	}
	fmt.Println("Database Admin migrated!")
	if err := db.AutoMigrate(&model.Service{}); err != nil {
		log.Fatal("migrate services: ", err)
	}
	fmt.Println("Database Service migrated!")
	if err := db.AutoMigrate(&model.NailTechnician{}); err != nil {
		log.Fatal("migrate nail technicians: ", err)
	}
	fmt.Println("Database Nail Technician migrated!")

	// Legacy schemas may have a composite primary key containing a custom string ID.
	// These unique indexes let booking foreign keys safely reference gorm.Model.ID.
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_service_dbs_gorm_id ON service_dbs (id)").Error; err != nil {
		log.Fatal("prepare service booking foreign key: ", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_nail_technician_dbs_gorm_id ON nail_technician_dbs (id)").Error; err != nil {
		log.Fatal("prepare technician booking foreign key: ", err)
	}
	if db.Migrator().HasTable(&model.Booking{}) {
		if err := db.Exec("ALTER TABLE bookings ALTER COLUMN user_id DROP NOT NULL").Error; err != nil {
			log.Fatal("make booking user optional: ", err)
		}
	}
	if err := db.AutoMigrate(&model.Booking{}); err != nil {
		log.Fatal("migrate bookings: ", err)
	}
	fmt.Println("Database Booking migrated!")
	if err := db.AutoMigrate(&model.ShopSetting{}); err != nil {
		log.Fatal("migrate shop settings: ", err)
	}
	fmt.Println("Database Shop Setting migrated!")

	r := router.New(db, cfg.AllowOrigin, jwtManager, customerJWTManager, cfg.LineLoginChannelID)
	r.Run(":" + cfg.Port)
}

func setThailandTimezone() {
	time.Local = time.FixedZone("Asia/Bangkok", 7*60*60)
}
