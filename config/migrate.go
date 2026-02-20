package config

import (
	"work-management-system/models"
)

func Migrate() {
    err := DB.AutoMigrate(
        // base/reference tables first
        &models.Role{},
        &models.Permission{},
        &models.RolePermission{},

        // other base tables
        &models.Customer{},

        // user first (many tables depend on it)
        &models.User{},

        // project after user (project often references user/customer)
        &models.Project{},
        &models.ProjectRequest{},

        // team structures
        &models.Team{},
        &models.TeamMember{},

        // tasks and task-related
        &models.Task{},
        &models.TaskReviewComment{},
        &models.Approval{},

        // files usually depend on project/task/user
        &models.File{},
        &models.ProjectFinalFile{},

        // misc dependent tables
        &models.CalendarEvent{},
        &models.Notification{},
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
