package controllers

import (
	"net/http"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminController struct {
	Repo         *repositories.AdminRepository
	ActivityRepo *repositories.ActivityLogRepository
}

func NewAdminController(repo *repositories.AdminRepository, activityRepo *repositories.ActivityLogRepository) *AdminController {
	return &AdminController{
		Repo:         repo,
		ActivityRepo: activityRepo,
	}
}

func (ac *AdminController) Index(c *gin.Context) {
	var totalUsers int64
	var activeUsers int64
	var totalProjects int64
	var activeProjects int64
	var completedProjects int64
	var pendingRequests int64
	var totalTasks int64
	var todoTasks int64
	var inProgressTasks int64
	var forReviewTasks int64
	var doneTasks int64
	var totalTeams int64
	var totalDepartments int64
	var pendingInitialApprovals int64
	var pendingFinalQA int64
	var unreadNotifications int64

	_ = ac.Repo.DB.Model(&models.User{}).Count(&totalUsers).Error
	_ = ac.Repo.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Count(&totalProjects).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Where("status = ?", "in_progress").Count(&activeProjects).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Where("status = ?", "completed").Count(&completedProjects).Error
	_ = ac.Repo.DB.Model(&models.ProjectRequest{}).Where("status = ?", "pending").Count(&pendingRequests).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Count(&totalTasks).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Where("status = ?", "todo").Count(&todoTasks).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Where("status = ?", "in_progress").Count(&inProgressTasks).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Where("status = ?", "for_review").Count(&forReviewTasks).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Where("status = ?", "done").Count(&doneTasks).Error
	_ = ac.Repo.DB.Model(&models.Team{}).Count(&totalTeams).Error
	_ = ac.Repo.DB.Model(&models.Department{}).Count(&totalDepartments).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Where("approval_status = ?", "pending_initial_approval").Count(&pendingInitialApprovals).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Where("final_qa_status = ?", "pending_qa_review").Count(&pendingFinalQA).Error

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err == nil {
		_ = ac.Repo.DB.Model(&models.Notification{}).
			Where("user_id = ? AND is_read = ?", userID, false).
			Count(&unreadNotifications).Error
	}

	recentProjects := make([]models.Project, 0)
	_ = ac.Repo.DB.
		Preload("Team").
		Preload("Manager").
		Order("updated_at DESC").
		Limit(6).
		Find(&recentProjects).Error

	recentRequests := make([]models.ProjectRequest, 0)
	_ = ac.Repo.DB.
		Preload("Customer.User").
		Where("status = ?", "pending").
		Order("created_at DESC").
		Limit(6).
		Find(&recentRequests).Error

	recentLogs, _ := ac.ActivityRepo.ListRecent(10)

	c.HTML(http.StatusOK, "admin/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":               "Admin Dashboard",
		"ActivePage":              "dashboard",
		"TotalUsers":              totalUsers,
		"ActiveUsers":             activeUsers,
		"TotalProjects":           totalProjects,
		"ActiveProjects":          activeProjects,
		"CompletedProjects":       completedProjects,
		"PendingRequests":         pendingRequests,
		"TotalTasks":              totalTasks,
		"TodoTasks":               todoTasks,
		"InProgressTasks":         inProgressTasks,
		"ForReviewTasks":          forReviewTasks,
		"DoneTasks":               doneTasks,
		"TotalTeams":              totalTeams,
		"TotalDepartments":        totalDepartments,
		"PendingInitialApprovals": pendingInitialApprovals,
		"PendingFinalQA":          pendingFinalQA,
		"UnreadNotifications":     unreadNotifications,
		"RecentProjects":          recentProjects,
		"RecentRequests":          recentRequests,
		"RecentLogs":              recentLogs,
	}))
}

func (ac *AdminController) Activities(c *gin.Context) {
	logs, err := ac.ActivityRepo.ListRecent(300)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load activity logs")
		return
	}

	c.HTML(http.StatusOK, "admin/activities/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Activity Logs",
		"ActivePage": "activities",
		"Logs":       logs,
	}))
}

func (ac *AdminController) Profile(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid admin ID")
		return
	}

	user, err := ac.Repo.GetUserByID(userID)
	if err != nil || user == nil {
		c.String(http.StatusNotFound, "Admin user not found")
		return
	}

	var totalUsers int64
	var activeUsers int64
	var totalProjects int64
	var totalTasks int64
	var totalTeams int64
	var totalDepartments int64

	_ = ac.Repo.DB.Model(&models.User{}).Count(&totalUsers).Error
	_ = ac.Repo.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers).Error
	_ = ac.Repo.DB.Model(&models.Project{}).Count(&totalProjects).Error
	_ = ac.Repo.DB.Model(&models.Task{}).Count(&totalTasks).Error
	_ = ac.Repo.DB.Model(&models.Team{}).Count(&totalTeams).Error
	_ = ac.Repo.DB.Model(&models.Department{}).Count(&totalDepartments).Error

	c.HTML(http.StatusOK, "admin/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":         "My Profile",
		"ActivePage":        "profile",
		"User":              user,
		"SystemUsers":       totalUsers,
		"ActiveUsers":       activeUsers,
		"SystemProjects":    totalProjects,
		"SystemTasks":       totalTasks,
		"SystemTeams":       totalTeams,
		"SystemDepartments": totalDepartments,
	}))
}

func (ac *AdminController) Notifications(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}

	notifications := make([]models.Notification, 0)
	if err := ac.Repo.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to load notifications")
		return
	}

	c.HTML(http.StatusOK, "admin/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
