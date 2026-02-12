package controllers

import (
	"net/http"
	"strings"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SupervisorController struct {
	ProjectRepo      *repositories.ProjectRepository
	UserRepo         *repositories.UserRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewSupervisorController(projectRepo *repositories.ProjectRepository, userRepo *repositories.UserRepository, notificationRepo *repositories.NotificationRepository) *SupervisorController {
	return &SupervisorController{
		ProjectRepo:      projectRepo,
		UserRepo:         userRepo,
		NotificationRepo: notificationRepo,
	}
}

func (sc *SupervisorController) Index(c *gin.Context) {
	supervisorID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid supervisor ID")
		return
	}

	projects, err := sc.ProjectRepo.GetBySupervisorID(supervisorID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load dashboard projects")
		return
	}

	projectIDs := make([]uuid.UUID, 0, len(projects))
	completedProjects := 0
	pendingFinalQA := 0
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
		if p.Status == "completed" || p.FinalQAStatus == "approved" {
			completedProjects++
		}
		if p.FinalQAStatus == "pending_qa_review" {
			pendingFinalQA++
		}
	}

	var totalTasks int64
	var doneTasks int64
	var inProgressTasks int64
	var forReviewTasks int64
	var todoTasks int64
	var teamMemberCount int64
	recentTasks := []models.Task{}

	if len(projectIDs) > 0 {
		_ = sc.ProjectRepo.DB.Model(&models.Task{}).
			Where("project_id IN ?", projectIDs).
			Count(&totalTasks).Error
		_ = sc.ProjectRepo.DB.Model(&models.Task{}).
			Where("project_id IN ? AND status = ?", projectIDs, "done").
			Count(&doneTasks).Error
		_ = sc.ProjectRepo.DB.Model(&models.Task{}).
			Where("project_id IN ? AND status = ?", projectIDs, "in_progress").
			Count(&inProgressTasks).Error
		_ = sc.ProjectRepo.DB.Model(&models.Task{}).
			Where("project_id IN ? AND status = ?", projectIDs, "for_review").
			Count(&forReviewTasks).Error
		_ = sc.ProjectRepo.DB.Model(&models.Task{}).
			Where("project_id IN ? AND status = ?", projectIDs, "todo").
			Count(&todoTasks).Error

		_ = sc.ProjectRepo.DB.
			Joins("JOIN projects ON projects.id = tasks.project_id").
			Where("projects.id IN ?", projectIDs).
			Preload("Project").
			Preload("Assignee").
			Order("tasks.updated_at DESC").
			Limit(8).
			Find(&recentTasks).Error

		_ = sc.ProjectRepo.DB.
			Table("team_members").
			Joins("JOIN teams ON teams.id = team_members.team_id").
			Where("teams.supervisor_id = ? AND team_members.user_id <> ?", supervisorID, supervisorID).
			Distinct("team_members.user_id").
			Count(&teamMemberCount).Error
	}

	var completionRate int64
	if totalTasks > 0 {
		completionRate = (doneTasks * 100) / totalTasks
	}

	c.HTML(http.StatusOK, "supervisor/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":         "Supervisor Dashboard",
		"ActivePage":        "dashboard",
		"TotalProjects":     len(projects),
		"CompletedProjects": completedProjects,
		"PendingFinalQA":    pendingFinalQA,
		"TeamMemberCount":   teamMemberCount,
		"TotalTasks":        totalTasks,
		"DoneTasks":         doneTasks,
		"InProgressTasks":   inProgressTasks,
		"ForReviewTasks":    forReviewTasks,
		"TodoTasks":         todoTasks,
		"CompletionRate":    completionRate,
		"RecentTasks":       recentTasks,
	}))
}

func (sc *SupervisorController) Profile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := sc.UserRepo.FindByID(userIDStr)
	if err != nil || user == nil {
		c.String(http.StatusNotFound, "User not found")
		return
	}

	var teams []models.Team
	_ = sc.ProjectRepo.DB.
		Where("supervisor_id = ?", user.ID).
		Preload("Members").
		Find(&teams).Error

	c.HTML(http.StatusOK, "supervisor/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "My Profile",
		"ActivePage": "profile",
		"User":       user,
		"Teams":      teams,
	}))
}

func (sc *SupervisorController) Projects(c *gin.Context) {
	supervisorID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid supervisor ID")
		return
	}

	projects, err := sc.ProjectRepo.GetBySupervisorID(supervisorID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load projects")
		return
	}

	finalFilesByProject := map[string][]gin.H{}
	if len(projects) > 0 {
		projectIDs := make([]uuid.UUID, 0, len(projects))
		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}

		var finalFiles []models.ProjectFinalFile
		_ = sc.ProjectRepo.DB.
			Where("project_id IN ?", projectIDs).
			Order("created_at DESC").
			Find(&finalFiles).Error

		for _, file := range finalFiles {
			key := file.ProjectID.String()
			downloadURL := "/" + strings.ReplaceAll(file.FilePath, "\\", "/")
			finalFilesByProject[key] = append(finalFilesByProject[key], gin.H{
				"id":           file.ID,
				"file_name":    file.FileName,
				"download_url": downloadURL,
				"uploaded_at":  file.CreatedAt,
			})
		}
	}

	c.HTML(http.StatusOK, "supervisor/projects/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":           "Team Projects",
		"ActivePage":          "projects",
		"projects":            projects,
		"FinalFilesByProject": finalFilesByProject,
	}))
}

func (sc *SupervisorController) SubmitProjectForQA(c *gin.Context) {
	supervisorID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid supervisor ID")
		return
	}

	projectID := c.Param("id")
	project, err := sc.ProjectRepo.GetByID(projectID)
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}

	if project.TeamID == nil {
		c.String(http.StatusForbidden, "Project is not assigned to a team")
		return
	}

	var canAccess int64
	if err := sc.ProjectRepo.DB.
		Table("teams").
		Where("id = ? AND supervisor_id = ?", *project.TeamID, supervisorID).
		Count(&canAccess).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to validate project access")
		return
	}
	if canAccess == 0 {
		c.String(http.StatusForbidden, "You can only submit your own team projects")
		return
	}

	if project.ApproverID == nil {
		c.String(http.StatusBadRequest, "Assign a QA approver to this project before submission")
		return
	}

	var finalFileCount int64
	if err := sc.ProjectRepo.DB.Model(&models.ProjectFinalFile{}).Where("project_id = ?", project.ID).Count(&finalFileCount).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify final project file")
		return
	}
	if finalFileCount == 0 {
		c.String(http.StatusBadRequest, "Upload at least one final project file before submitting to QA")
		return
	}

	var totalTasks int64
	if err := sc.ProjectRepo.DB.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&totalTasks).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to count project tasks")
		return
	}

	var doneTasks int64
	if err := sc.ProjectRepo.DB.Model(&models.Task{}).Where("project_id = ? AND status = ?", project.ID, "done").Count(&doneTasks).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify task completion")
		return
	}

	if totalTasks > 0 && doneTasks != totalTasks {
		c.String(http.StatusBadRequest, "All project tasks must be done before final QA submission")
		return
	}

	if project.FinalQAStatus == "pending_qa_review" {
		c.String(http.StatusBadRequest, "Project is already pending final QA review")
		return
	}

	now := time.Now()
	if err := sc.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"final_qa_status":       "pending_qa_review",
			"final_qa_submitted_at": &now,
			"final_qa_submitted_by": supervisorID,
			"final_qa_reviewed_at":  nil,
			"final_qa_reviewed_by":  nil,
			"final_qa_comment":      "",
			"status":                "completed",
		}).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to submit project for QA")
		return
	}
	_ = sc.ProjectRepo.SyncWorkflowStatus(project.ID)

	_ = sc.NotificationRepo.CreateForUserWithLink(*project.ApproverID, "Project submitted for final QA: "+project.Name, "/qa/projects/reviews")

	c.Redirect(http.StatusSeeOther, "/supervisor/projects")
}

func (sc *SupervisorController) Notifications(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}

	notifications, err := sc.NotificationRepo.RecentByUser(userID, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load notifications")
		return
	}

	c.HTML(http.StatusOK, "supervisor/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
