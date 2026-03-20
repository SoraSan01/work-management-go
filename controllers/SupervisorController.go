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
			Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
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

	allowedUsers := make(map[uuid.UUID]models.User)
	for _, project := range projects {
		for _, member := range project.Team.Members {
			if member.ID == supervisorID {
				continue
			}
			allowedUsers[member.ID] = member
		}
	}

	users := make([]models.User, 0, len(allowedUsers))
	for _, user := range allowedUsers {
		users = append(users, user)
	}

	c.HTML(http.StatusOK, "supervisor/projects/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Team Projects",
		"ActivePage": "projects",
		"projects":   projects,
		"Users":      users,
	}))
}

func (sc *SupervisorController) AssignProjectEmployee(c *gin.Context) {
	supervisorID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid supervisor ID")
		return
	}

	projectID := c.Param("id")
	assignedEmployeeIDStr := strings.TrimSpace(c.PostForm("assigned_employee_id"))
	priority := normalizePriority(c.PostForm("priority"))
	if assignedEmployeeIDStr == "" {
		c.String(http.StatusBadRequest, "Assigned employee is required")
		return
	}

	project, err := sc.ProjectRepo.GetByID(projectID)
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}

	if project.TeamID == nil {
		c.String(http.StatusBadRequest, "Project is not assigned to a team")
		return
	}

	var teamCount int64
	if err := sc.ProjectRepo.DB.
		Table("teams").
		Where("id = ? AND supervisor_id = ?", *project.TeamID, supervisorID).
		Count(&teamCount).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to validate project access")
		return
	}
	if teamCount == 0 {
		c.String(http.StatusForbidden, "You can only assign employees for your own team projects")
		return
	}

	assignedEmployeeID, err := uuid.Parse(assignedEmployeeIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid assigned employee ID")
		return
	}

	if assignedEmployeeID == project.Team.SupervisorID {
		c.String(http.StatusBadRequest, "Supervisors cannot assign projects to themselves")
		return
	}

	var memberCount int64
	if err := sc.ProjectRepo.DB.
		Table("team_members").
		Where("team_id = ? AND user_id = ?", *project.TeamID, assignedEmployeeID).
		Count(&memberCount).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to validate assigned employee")
		return
	}
	if memberCount == 0 {
		c.String(http.StatusBadRequest, "Assigned employee must belong to your team")
		return
	}

	if err := sc.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"assigned_employee_id": assignedEmployeeID,
			"priority":             priority,
			"updated_at":           time.Now(),
		}).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to assign employee to project")
		return
	}

	project.AssignedEmployeeID = &assignedEmployeeID
	project.Priority = priority
	_ = sc.ProjectRepo.SyncTaskAssignees(project.ID, assignedEmployeeID)
	createdKickoffTask, err := sc.ProjectRepo.EnsureKickoffTask(&project)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to prepare employee taskboard")
		return
	}
	_ = sc.ProjectRepo.SyncWorkflowStatus(project.ID)

	notification := "A project was assigned to you: " + project.Name
	if createdKickoffTask {
		notification = "A new project is ready for you to start: " + project.Name
	}
	_ = sc.NotificationRepo.CreateForUserWithLink(assignedEmployeeID, notification, "/employee/tasks/board")

	c.Redirect(http.StatusSeeOther, "/supervisor/projects")
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

	now := time.Now()
	if err := sc.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"approval_status": "pending_final_qa",
			"status":          "completed",
			"updated_at":      now,
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
