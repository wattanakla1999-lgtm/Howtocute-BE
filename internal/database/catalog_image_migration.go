package database

import "gorm.io/gorm"

func EnsureCatalogImageSchema(db *gorm.DB) error {
	for _, table := range []string{"service_dbs", "services"} {
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Exec(`
			ALTER TABLE ` + table + `
			ADD COLUMN IF NOT EXISTS image_url TEXT,
			ADD COLUMN IF NOT EXISTS img TEXT
		`).Error; err != nil {
			return err
		}
		if db.Migrator().HasColumn(table, "service_img") {
			if err := db.Exec(`
				UPDATE ` + table + `
				SET img = COALESCE(NULLIF(img, ''), service_img),
					image_url = COALESCE(NULLIF(image_url, ''), service_img)
				WHERE service_img IS NOT NULL AND service_img <> ''
			`).Error; err != nil {
				return err
			}
		}
	}

	for _, table := range []string{"nail_technician_dbs", "technicians"} {
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Exec(`
			ALTER TABLE ` + table + `
			ADD COLUMN IF NOT EXISTS profile_img TEXT,
			ADD COLUMN IF NOT EXISTS avatar_url TEXT
		`).Error; err != nil {
			return err
		}
	}

	return nil
}
