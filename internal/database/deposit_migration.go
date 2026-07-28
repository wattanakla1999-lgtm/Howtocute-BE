package database

import "gorm.io/gorm"

func EnsureDepositPaymentSchema(db *gorm.DB) error {
	if db.Migrator().HasTable("shop_settings") {
		if err := db.Exec(`
			ALTER TABLE shop_settings
			ADD COLUMN IF NOT EXISTS prompt_pay_number VARCHAR(20) DEFAULT '0812345678',
			ADD COLUMN IF NOT EXISTS account_name VARCHAR(100) DEFAULT 'ร้าน Nailly Nail Salon',
			ADD COLUMN IF NOT EXISTS bank_name VARCHAR(50) DEFAULT 'ธนาคารกสิกรไทย',
			ADD COLUMN IF NOT EXISTS deposit_amount NUMERIC(10,2) DEFAULT 200.00,
			ADD COLUMN IF NOT EXISTS qr_code_url TEXT DEFAULT '',
			ADD COLUMN IF NOT EXISTS account_number VARCHAR(50) DEFAULT ''
		`).Error; err != nil {
			return err
		}
		if err := db.Exec(`
			UPDATE shop_settings
			SET prompt_pay_number = COALESCE(NULLIF(prompt_pay_number, ''), '0812345678'),
				account_name = COALESCE(NULLIF(account_name, ''), 'ร้าน Nailly Nail Salon'),
				bank_name = COALESCE(NULLIF(bank_name, ''), 'ธนาคารกสิกรไทย'),
				deposit_amount = COALESCE(deposit_amount, 200.00)
		`).Error; err != nil {
			return err
		}
	}

	if db.Migrator().HasTable("bookings") {
		if err := db.Exec(`
			ALTER TABLE bookings
			ADD COLUMN IF NOT EXISTS technician_id BIGINT,
			ADD COLUMN IF NOT EXISTS deposit_amount NUMERIC(10,2) DEFAULT 0.00,
			ADD COLUMN IF NOT EXISTS deposit_status VARCHAR(20) DEFAULT 'none',
			ADD COLUMN IF NOT EXISTS slip_url TEXT,
			ADD COLUMN IF NOT EXISTS slip_uploaded_at TIMESTAMP WITH TIME ZONE,
			ADD COLUMN IF NOT EXISTS deposit_reject_reason TEXT
		`).Error; err != nil {
			return err
		}
		if err := db.Exec(`
			UPDATE bookings
			SET deposit_amount = COALESCE(deposit_amount, 0.00),
				deposit_status = COALESCE(NULLIF(deposit_status, ''), 'none')
		`).Error; err != nil {
			return err
		}
	}

	return nil
}
