package config

import (
	"log"
	"strings"

	"work-management-system/models"
)

func Migrate() {
	prepareUserManagerConstraint()

	err := DB.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.Customer{},
		&models.User{},
		&models.Project{},
		&models.ProjectRequest{},
		&models.Team{},
		&models.TeamMember{},
		&models.Task{},
		&models.TaskReviewComment{},
		&models.Approval{},
		&models.File{},
		&models.CalendarEvent{},
		&models.Notification{},
		&models.ActivityLog{},
	)
	if err != nil {
		panic(err)
	}

	dropObsoleteTables("project_final_files")

	repairUserManagerConstraint()
}

func dropObsoleteTables(tables ...string) {
	for _, table := range tables {
		if !DB.Migrator().HasTable(table) {
			continue
		}

		if err := DB.Migrator().DropTable(table); err != nil {
			panic(err)
		}

		log.Printf("Dropped obsolete table: %s", table)
	}
}

func dropObsoleteColumns(columnsByModel map[any][]string) {
	for model, columns := range columnsByModel {
		for _, column := range columns {
			if !DB.Migrator().HasColumn(model, column) {
				continue
			}

			if err := DB.Migrator().DropColumn(model, column); err != nil {
				panic(err)
			}

			log.Printf("Dropped obsolete column: %T.%s", model, column)
		}
	}
}

func repairUserManagerConstraint() {
	type constraintInfo struct {
		Name       string
		Definition string
	}

	var constraints []constraintInfo
	if err := DB.Raw(`
		SELECT conname AS name, pg_get_constraintdef(oid) AS definition
		FROM pg_constraint
		WHERE conrelid = 'users'::regclass
		  AND conname IN ('fk_projects_manager', 'fk_users_manager')
	`).Scan(&constraints).Error; err != nil {
		panic(err)
	}

	for _, constraint := range constraints {
		if constraint.Name == "fk_projects_manager" && strings.Contains(constraint.Definition, "REFERENCES projects") {
			if err := DB.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_projects_manager`).Error; err != nil {
				panic(err)
			}
			log.Printf("Dropped invalid users.manager_id constraint: %s", constraint.Name)
		}
	}

	var validCount int64
	if err := DB.Raw(`
		SELECT COUNT(*)
		FROM pg_constraint
		WHERE conrelid = 'users'::regclass
		  AND contype = 'f'
		  AND conkey = ARRAY[
			(SELECT attnum FROM pg_attribute WHERE attrelid = 'users'::regclass AND attname = 'manager_id')
		  ]::smallint[]
		  AND confrelid = 'users'::regclass
	`).Scan(&validCount).Error; err != nil {
		panic(err)
	}

	if validCount == 0 {
		if err := DB.Exec(`
			ALTER TABLE users
			ADD CONSTRAINT fk_users_manager
			FOREIGN KEY (manager_id) REFERENCES users(id)
			ON UPDATE CASCADE
			ON DELETE SET NULL
		`).Error; err != nil {
			panic(err)
		}
		log.Printf("Created valid users.manager_id foreign key")
	}
}

func prepareUserManagerConstraint() {
	if !DB.Migrator().HasTable(&models.User{}) || !DB.Migrator().HasColumn(&models.User{}, "manager_id") {
		return
	}

	if err := DB.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_manager`).Error; err != nil {
		panic(err)
	}

	type constraintInfo struct {
		Name       string
		Definition string
	}

	var constraints []constraintInfo
	if err := DB.Raw(`
		SELECT conname AS name, pg_get_constraintdef(oid) AS definition
		FROM pg_constraint
		WHERE conrelid = 'users'::regclass
		  AND conname = 'fk_projects_manager'
	`).Scan(&constraints).Error; err != nil {
		panic(err)
	}

	for _, constraint := range constraints {
		if strings.Contains(constraint.Definition, "REFERENCES projects") {
			if err := DB.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_projects_manager`).Error; err != nil {
				panic(err)
			}
			log.Printf("Dropped invalid pre-migrate users.manager_id constraint: %s", constraint.Name)
		}
	}

	var validCount int64
	if err := DB.Raw(`
		SELECT COUNT(*)
		FROM pg_constraint
		WHERE conrelid = 'users'::regclass
		  AND conname = 'fk_projects_manager'
		  AND contype = 'f'
		  AND confrelid = 'users'::regclass
	`).Scan(&validCount).Error; err != nil {
		panic(err)
	}

	if validCount == 0 {
		if err := DB.Exec(`
			ALTER TABLE users
			ADD CONSTRAINT fk_projects_manager
			FOREIGN KEY (manager_id) REFERENCES users(id)
			ON UPDATE CASCADE
			ON DELETE SET NULL
		`).Error; err != nil {
			panic(err)
		}
		log.Printf("Prepared valid users.manager_id foreign key for migration")
	}
}
