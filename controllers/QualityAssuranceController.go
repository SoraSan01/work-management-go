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

type QualityAssuranceController struct {
	Repo             *repositories.QualityAssuranceRepository
	UserRepo         *repositories.UserRepository
	TaskRepo         *repositories.TaskRepository
	ProjectRepo      *repositories.ProjectRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewQualityAssuranceController(
	repo *repositories.QualityAssuranceRepository,
	userRepo *repositories.UserRepository,
	taskRepo *repositories.TaskRepository,
	projectRepo *repositories.ProjectRepository,
	notificationRepo *repositories.NotificationRepository,
) *QualityAssuranceController {
	return &QualityAssuranceController{
		Repo:             repo,
		UserRepo:         userRepo,
		TaskRepo:         taskRepo,
		ProjectRepo:      projectRepo,
		NotificationRepo: notificationRepo,
	}
}

func (ac *QualityAssuranceController) Index(c *gin.Context) {
	qaID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid quality assurance ID")
		return
	}

	projects, err := ac.ProjectRepo.GetByAssignedQAID(qaID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assigned projects")
		return
	}

	tasks, _ := ac.TaskRepo.GetByQATeamMemberID(qaID)

	assignedProjects := int64(len(projects))
	var pendingFinalQA int64
	var approvedProjects int64
	var sentProjects int64
	for _, p := range projects {
		if p.FinalQAStatus == "pending_qa_review" {
			pendingFinalQA++
		}
		if p.FinalQAStatus == "approved" {
			approvedProjects++
			if p.FinalQASent {
				sentProjects++
			}
		}
	}

	var forReviewTasks int64
	var doneTasks int64
	for _, t := range tasks {
		if t.Status == "for_review" {
			forReviewTasks++
		}
		if t.Status == "done" {
			doneTasks++
		}
	}

	c.HTML(http.StatusOK, "qa/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Quality Assurance Dashboard",
		"ActivePage":       "dashboard",
		"AssignedProjects": assignedProjects,
		"PendingFinalQA":   pendingFinalQA,
		"ApprovedProjects": approvedProjects,
		"SentProjects":     sentProjects,
		"ForReviewTasks":   forReviewTasks,
		"DoneTasks":        doneTasks,
	}))
}

func (ac *QualityAssuranceController) ProjectReviews(c *gin.Context) {
	qaID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid quality assurance ID")
		return
	}

	projects, err := ac.ProjectRepo.GetByAssignedQAID(qaID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assigned projects")
		return
	}

	finalFilesByProject := map[string][]gin.H{}
	if len(projects) > 0 {
		projectIDs := make([]uuid.UUID, 0, len(projects))
		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}

		var finalFiles []models.ProjectFinalFile
		_ = ac.ProjectRepo.DB.
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

	c.HTML(http.StatusOK, "qa/projects/reviews.html", utils.TemplateContext(c, gin.H{
		"PageTitle":           "Project Final Review",
		"ActivePage":          "project-reviews",
		"projects":            projects,
		"FinalFilesByProject": finalFilesByProject,
	}))
}

func (ac *QualityAssuranceController) ProcessProjectReview(c *gin.Context) {
	projectID := c.Param("id")
	qaID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid quality assurance ID")
		return
	}

	var form struct {
		Decision string `form:"decision"`
		Comment  string `form:"comment"`
	}
	_ = c.ShouldBind(&form)

	decision := strings.ToLower(strings.TrimSpace(form.Decision))
	if decision == "" {
		decision = strings.ToLower(strings.TrimSpace(c.PostForm("decision")))
	}
	if decision != "approved" && decision != "rejected" {
		c.String(http.StatusBadRequest, "Decision must be approved or rejected")
		return
	}

	project, err := ac.ProjectRepo.GetByID(projectID)
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}

	if project.ApproverID == nil || *project.ApproverID != qaID {
		c.String(http.StatusForbidden, "You can only review projects assigned to you")
		return
	}

	if project.FinalQAStatus != "pending_qa_review" {
		c.String(http.StatusBadRequest, "Project is not awaiting final QA review")
		return
	}

	sendNow := strings.EqualFold(strings.TrimSpace(c.PostForm("send_now")), "true")
	now := time.Now()
	newProjectStatus := "in_progress"
	if decision == "approved" && sendNow {
		newProjectStatus = "completed"
	}

	updates := map[string]interface{}{
		"final_qa_status":      decision,
		"final_qa_comment":     strings.TrimSpace(c.PostForm("comment")),
		"final_qa_reviewed_by": qaID,
		"final_qa_reviewed_at": &now,
		"status":               newProjectStatus,
	}
	if decision == "approved" {
		updates["final_qa_sent"] = sendNow
		if sendNow {
			updates["final_qa_sent_at"] = &now
			updates["final_qa_sent_by"] = qaID
		} else {
			updates["final_qa_sent_at"] = nil
			updates["final_qa_sent_by"] = nil
		}
	} else {
		updates["final_qa_sent"] = false
		updates["final_qa_sent_at"] = nil
		updates["final_qa_sent_by"] = nil
	}

	err = ac.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(updates).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update project review")
		return
	}
	_ = ac.ProjectRepo.SyncWorkflowStatus(project.ID)

	if project.TeamID != nil {
		var team models.Team
		if teamErr := ac.ProjectRepo.DB.Select("supervisor_id").Where("id = ?", *project.TeamID).First(&team).Error; teamErr == nil {
			_ = ac.NotificationRepo.CreateForUserWithLink(team.SupervisorID, "Final QA "+decision+" project: "+project.Name, "/supervisor/projects")
		}
	}

	c.Redirect(http.StatusSeeOther, "/qa/projects/reviews")
}

func (ac *QualityAssuranceController) SendApprovedProject(c *gin.Context) {
	projectID := c.Param("id")
	qaID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid quality assurance ID")
		return
	}

	project, err := ac.ProjectRepo.GetByID(projectID)
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}
	if project.ApproverID == nil || *project.ApproverID != qaID {
		c.String(http.StatusForbidden, "You can only send projects assigned to you")
		return
	}
	if project.FinalQAStatus != "approved" {
		c.String(http.StatusBadRequest, "Only approved projects can be sent")
		return
	}
	if project.FinalQASent {
		c.Redirect(http.StatusSeeOther, "/qa/projects/reviews")
		return
	}

	now := time.Now()
	if err := ac.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"final_qa_sent":    true,
			"final_qa_sent_at": &now,
			"final_qa_sent_by": qaID,
			"status":           "completed",
		}).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to send project")
		return
	}
	_ = ac.ProjectRepo.SyncWorkflowStatus(project.ID)

	if project.TeamID != nil {
		var team models.Team
		if teamErr := ac.ProjectRepo.DB.Select("supervisor_id").Where("id = ?", *project.TeamID).First(&team).Error; teamErr == nil {
			_ = ac.NotificationRepo.CreateForUserWithLink(team.SupervisorID, "Final QA sent approved project: "+project.Name, "/supervisor/projects")
		}
	}

	c.Redirect(http.StatusSeeOther, "/qa/projects/reviews")
}

func (ac *QualityAssuranceController) Profile(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := ac.UserRepo.FindByID(userIDStr)
	if err != nil || user == nil {
		c.String(http.StatusNotFound, "User not found")
		return
	}

	qaID, _ := uuid.Parse(userIDStr)
	projects, _ := ac.ProjectRepo.GetByAssignedQAID(qaID)
	var approvedCount int64
	var sentCount int64
	for _, p := range projects {
		if p.FinalQAStatus == "approved" {
			approvedCount++
			if p.FinalQASent {
				sentCount++
			}
		}
	}

	c.HTML(http.StatusOK, "qa/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "My Profile",
		"ActivePage":       "profile",
		"User":             user,
		"AssignedProjects": len(projects),
		"ApprovedProjects": approvedCount,
		"SentProjects":     sentCount,
	}))
}

func (ac *QualityAssuranceController) Notifications(c *gin.Context) {
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

	c.HTML(http.StatusOK, "qa/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
