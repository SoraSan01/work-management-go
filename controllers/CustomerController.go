package controllers

import (
	"net/http"
	"time"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomerController struct {
	Repo             *repositories.CustomerRepository
	UserRepo         *repositories.UserRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewCustomerController(
	repo *repositories.CustomerRepository,
	userRepo *repositories.UserRepository,
	notificationRepo *repositories.NotificationRepository,
) *CustomerController {
	return &CustomerController{
		Repo:             repo,
		UserRepo:         userRepo,
		NotificationRepo: notificationRepo,
	}
}

func (ac *CustomerController) Index(c *gin.Context) {
	customerID, err := uuid.Parse(c.GetString("customer_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid customer")
		return
	}

	type dashboardRequest struct {
		ID          uuid.UUID
		Title       string
		Status      string
		Deadline    time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
		ProjectName string
	}

	var recentRequests []dashboardRequest
	_ = ac.Repo.DB.
		Table("project_requests").
		Select("project_requests.id, project_requests.title, project_requests.status, project_requests.deadline, project_requests.created_at, project_requests.updated_at, projects.name as project_name").
		Joins("LEFT JOIN projects ON projects.project_request_id = project_requests.id AND projects.deleted_at IS NULL").
		Where("project_requests.customer_id = ?", customerID).
		Order("project_requests.updated_at DESC").
		Limit(6).
		Scan(&recentRequests).Error

	var totalRequests int64
	var pendingRequests int64
	var approvedRequests int64
	var rejectedRequests int64

	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ?", customerID).Count(&totalRequests).Error
	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ? AND status = ?", customerID, "pending").Count(&pendingRequests).Error
	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ? AND status = ?", customerID, "approved").Count(&approvedRequests).Error
	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ? AND status = ?", customerID, "rejected").Count(&rejectedRequests).Error

	var totalProjects int64
	var inProgressProjects int64
	var completedProjects int64
	var pendingFinalQA int64

	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("project_requests.customer_id = ? AND projects.deleted_at IS NULL", customerID).
		Count(&totalProjects).Error
	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("project_requests.customer_id = ? AND projects.status = ? AND projects.deleted_at IS NULL", customerID, "in_progress").
		Count(&inProgressProjects).Error
	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where(`
			project_requests.customer_id = ?
			AND projects.deleted_at IS NULL
			AND (
				projects.approval_status = ?
				OR projects.status = ?
				OR (
					EXISTS (SELECT 1 FROM tasks WHERE tasks.project_id = projects.id)
					AND NOT EXISTS (
						SELECT 1
						FROM tasks
						WHERE tasks.project_id = projects.id
						  AND tasks.status <> ?
					)
				)
			)
		`, customerID, "delivered", "completed", "done").
		Count(&completedProjects).Error
	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("project_requests.customer_id = ? AND projects.approval_status = ? AND projects.deleted_at IS NULL", customerID, "pending_final_qa").
		Count(&pendingFinalQA).Error

	var approvalRate int64
	if totalRequests > 0 {
		approvalRate = (approvedRequests * 100) / totalRequests
	}

	c.HTML(http.StatusOK, "customer/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":          "Customer Dashboard",
		"ActivePage":         "dashboard",
		"TotalRequests":      totalRequests,
		"PendingRequests":    pendingRequests,
		"ApprovedRequests":   approvedRequests,
		"RejectedRequests":   rejectedRequests,
		"TotalProjects":      totalProjects,
		"InProgressProjects": inProgressProjects,
		"CompletedProjects":  completedProjects,
		"PendingFinalQA":     pendingFinalQA,
		"ApprovalRate":       approvalRate,
		"RecentRequests":     recentRequests,
	}))
}

func (ac *CustomerController) Profile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	customerIDStr := c.GetString("customer_id")
	if userIDStr == "" || customerIDStr == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := ac.UserRepo.FindByID(userIDStr)
	if err != nil || user == nil {
		c.String(http.StatusNotFound, "User not found")
		return
	}

	customer, err := ac.Repo.FindByUserID(userIDStr)
	if err != nil || customer == nil {
		c.String(http.StatusNotFound, "Customer profile not found")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid customer")
		return
	}

	var totalRequests int64
	var approvedRequests int64
	var pendingRequests int64
	var totalProjects int64
	var completedProjects int64

	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ?", customerID).Count(&totalRequests).Error
	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ? AND status = ?", customerID, "approved").Count(&approvedRequests).Error
	_ = ac.Repo.DB.Table("project_requests").Where("customer_id = ? AND status = ?", customerID, "pending").Count(&pendingRequests).Error
	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("project_requests.customer_id = ? AND projects.deleted_at IS NULL", customerID).
		Count(&totalProjects).Error
	_ = ac.Repo.DB.
		Table("projects").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where(`
			project_requests.customer_id = ?
			AND projects.deleted_at IS NULL
			AND (
				projects.approval_status = ?
				OR projects.status = ?
				OR (
					EXISTS (SELECT 1 FROM tasks WHERE tasks.project_id = projects.id)
					AND NOT EXISTS (
						SELECT 1
						FROM tasks
						WHERE tasks.project_id = projects.id
						  AND tasks.status <> ?
					)
				)
			)
		`, customerID, "delivered", "completed", "done").
		Count(&completedProjects).Error

	c.HTML(http.StatusOK, "customer/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":         "My Profile",
		"ActivePage":        "profile",
		"User":              user,
		"Customer":          customer,
		"TotalRequests":     totalRequests,
		"ApprovedRequests":  approvedRequests,
		"PendingRequests":   pendingRequests,
		"TotalProjects":     totalProjects,
		"CompletedProjects": completedProjects,
	}))
}

func (ac *CustomerController) Notifications(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}

	notifications, err := ac.NotificationRepo.RecentByUser(userID, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load notifications")
		return
	}

	c.HTML(http.StatusOK, "customer/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
