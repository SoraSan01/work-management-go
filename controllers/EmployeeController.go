package controllers

import (
	"net/http"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeController struct {
	Repo             *repositories.EmployeeRepository
	UserRepo         *repositories.UserRepository
	TeamRepo         *repositories.TeamRepository
	TaskRepo         *repositories.TaskRepository
	ProjectRepo      *repositories.ProjectRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewEmployeeController(
	repo *repositories.EmployeeRepository,
	userRepo *repositories.UserRepository,
	teamRepo *repositories.TeamRepository,
	taskRepo *repositories.TaskRepository,
	projectRepo *repositories.ProjectRepository,
	notificationRepo *repositories.NotificationRepository,
) *EmployeeController {
	return &EmployeeController{
		Repo:             repo,
		UserRepo:         userRepo,
		TeamRepo:         teamRepo,
		TaskRepo:         taskRepo,
		ProjectRepo:      projectRepo,
		NotificationRepo: notificationRepo,
	}
}

func (ec *EmployeeController) Index(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid employee ID")
		return
	}

	tasks, err := ec.TaskRepo.GetByAssignee(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load tasks")
		return
	}

	var todoTasks int64
	var inProgressTasks int64
	var forReviewTasks int64
	var doneTasks int64
	var completedThisWeek int64

	recentTasks := make([]models.Task, 0, 6)
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)

	for _, task := range tasks {
		switch task.Status {
		case "todo":
			todoTasks++
		case "in_progress":
			inProgressTasks++
		case "for_review":
			forReviewTasks++
		case "done":
			doneTasks++
			if task.UpdatedAt.After(weekStart) {
				completedThisWeek++
			}
		}
	}

	if len(tasks) > 0 {
		recentTasks = append(recentTasks, tasks...)
		// simple in-memory sort by updated_at desc
		for i := 0; i < len(recentTasks); i++ {
			for j := i + 1; j < len(recentTasks); j++ {
				if recentTasks[j].UpdatedAt.After(recentTasks[i].UpdatedAt) {
					recentTasks[i], recentTasks[j] = recentTasks[j], recentTasks[i]
				}
			}
		}
		if len(recentTasks) > 6 {
			recentTasks = recentTasks[:6]
		}
	}

	teams, _ := ec.TeamRepo.GetByUserID(userID)
	teamName := "-"
	if len(teams) > 0 {
		teamName = teams[0].Name
	}

	totalTasks := int64(len(tasks))
	openTasks := totalTasks - doneTasks
	if openTasks < 0 {
		openTasks = 0
	}
	var completionRate int64
	if totalTasks > 0 {
		completionRate = (doneTasks * 100) / totalTasks
	}

	c.HTML(http.StatusOK, "employee/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":         "Employee Dashboard",
		"ActivePage":        "dashboard",
		"TotalTasks":        totalTasks,
		"TodoTasks":         todoTasks,
		"InProgressTasks":   inProgressTasks,
		"ForReviewTasks":    forReviewTasks,
		"DoneTasks":         doneTasks,
		"OpenTasks":         openTasks,
		"CompletionRate":    completionRate,
		"CompletedThisWeek": completedThisWeek,
		"RecentTasks":       recentTasks,
		"TeamName":          teamName,
	}))
}

func (ec *EmployeeController) Profile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := ec.UserRepo.FindByID(userIDStr)
	if err != nil || user == nil {
		c.String(http.StatusNotFound, "User not found")
		return
	}

	userID, _ := uuid.Parse(userIDStr)
	teams, _ := ec.TeamRepo.GetByUserID(userID)
	tasks, _ := ec.TaskRepo.GetByAssignee(userID)

	var doneTasks int64
	for _, t := range tasks {
		if t.Status == "done" {
			doneTasks++
		}
	}

	c.HTML(http.StatusOK, "employee/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "My Profile",
		"ActivePage": "profile",
		"User":       user,
		"Teams":      teams,
		"TotalTasks": len(tasks),
		"DoneTasks":  doneTasks,
		"InProgress": len(tasks) - int(doneTasks),
	}))
}

func (ec *EmployeeController) Notifications(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}

	notifications, err := ec.NotificationRepo.RecentByUser(userID, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load notifications")
		return
	}

	c.HTML(http.StatusOK, "employee/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
