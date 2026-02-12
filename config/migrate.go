package config

import (
	"work-management-system/models"
)

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.TeamMember{},
		&models.Team{},
		&models.Task{},
		&models.RolePermission{},
		&models.Role{},
		&models.ProjectRequest{},
		&models.Project{},
		&models.Permission{},
		&models.Notification{},
		&models.File{},
		&models.ProjectFinalFile{},
		&models.Customer{},
		&models.CalendarEvent{},
		&models.TaskReviewComment{},
		&models.Approval{},
		&models.ActivityLog{},
	)
	if err != nil {
		panic(err)
	}

	if DB.Migrator().HasTable("comments") {
		if err := DB.Migrator().DropTable("comments"); err != nil {
			panic(err)
		}
	}
}
