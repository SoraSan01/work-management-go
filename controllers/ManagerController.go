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

type ManagerController struct {
	Repo             *repositories.ManagerRepository
	ProjectRepo      *repositories.ProjectRepository
	TeamRepo         *repositories.TeamRepository
	UserRepo         *repositories.UserRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewManagerController(
	repo *repositories.ManagerRepository,
	projectRepo *repositories.ProjectRepository,
	teamRepo *repositories.TeamRepository,
	userRepo *repositories.UserRepository,
	notificationRepo *repositories.NotificationRepository,
) *ManagerController {
	return &ManagerController{
		Repo:             repo,
		ProjectRepo:      projectRepo,
		TeamRepo:         teamRepo,
		UserRepo:         userRepo,
		NotificationRepo: notificationRepo,
	}
}

func (ac *ManagerController) Index(c *gin.Context) {
	managerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
		return
	}

	projects, err := ac.ProjectRepo.GetByManagerID(managerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load manager dashboard")
		return
	}

	projectIDs := make([]uuid.UUID, 0, len(projects))
	var activeProjects int64
	var completedProjects int64
	var pendingReviewProjects int64

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)

		if project.Status == "completed" {
			completedProjects++
		} else if project.Status != "cancelled" {
			activeProjects++
		}

		if project.ApprovalStatus == "pending_initial_approval" {
			pendingReviewProjects++
		}
	}

	var totalTasks int64
	var todoTasks int64
	var inProgressTasks int64
	var forReviewTasks int64
	var doneTasks int64

	if len(projectIDs) > 0 {
		_ = ac.ProjectRepo.DB.Model(&models.Task{}).Where("project_id IN ?", projectIDs).Count(&totalTasks).Error
		_ = ac.ProjectRepo.DB.Model(&models.Task{}).Where("project_id IN ? AND status = ?", projectIDs, "todo").Count(&todoTasks).Error
		_ = ac.ProjectRepo.DB.Model(&models.Task{}).Where("project_id IN ? AND status = ?", projectIDs, "in_progress").Count(&inProgressTasks).Error
		_ = ac.ProjectRepo.DB.Model(&models.Task{}).Where("project_id IN ? AND status = ?", projectIDs, "for_review").Count(&forReviewTasks).Error
		_ = ac.ProjectRepo.DB.Model(&models.Task{}).Where("project_id IN ? AND status = ?", projectIDs, "done").Count(&doneTasks).Error
	}

	var completionRate int64
	if totalTasks > 0 {
		completionRate = (doneTasks * 100) / totalTasks
	}

	pendingRequests, _ := ac.ProjectRepo.GetPendingRequests()
	recentRequests := pendingRequests
	if len(recentRequests) > 5 {
		recentRequests = recentRequests[:5]
	}

	recentProjects := projects
	if len(recentProjects) > 6 {
		recentProjects = recentProjects[:6]
	}

	c.HTML(http.StatusOK, "manager/dashboard/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":             "Manager Dashboard",
		"ActivePage":            "dashboard",
		"TotalProjects":         len(projects),
		"ActiveProjects":        activeProjects,
		"CompletedProjects":     completedProjects,
		"PendingReviewProjects": pendingReviewProjects,
		"PendingRequests":       len(pendingRequests),
		"TotalTasks":            totalTasks,
		"TodoTasks":             todoTasks,
		"InProgressTasks":       inProgressTasks,
		"ForReviewTasks":        forReviewTasks,
		"DoneTasks":             doneTasks,
		"CompletionRate":        completionRate,
		"RecentProjects":        recentProjects,
		"RecentRequests":        recentRequests,
	}))
}

func (ac *ManagerController) Profile(c *gin.Context) {
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

	managerID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
		return
	}

	projects, _ := ac.ProjectRepo.GetByManagerID(managerID)

	managedTeams := make([]models.Team, 0)
	if err := ac.TeamRepo.DB.
		Joins("JOIN users ON users.id = teams.supervisor_id").
		Where("users.manager_id = ?", managerID).
		Preload("Supervisor").
		Preload("Members").
		Order("teams.created_at DESC").
		Find(&managedTeams).Error; err != nil {
		managedTeams = []models.Team{}
	}

	var directReports int64
	_ = ac.UserRepo.DB.
		Model(&models.User{}).
		Where("manager_id = ? AND is_active = ?", managerID, true).
		Count(&directReports).Error

	var approvedProjects int64
	var rejectedProjects int64
	var inProgressProjects int64
	for _, project := range projects {
		switch project.ApprovalStatus {
		case "approved":
			approvedProjects++
		case "rejected":
			rejectedProjects++
		}
		if project.Status == "in_progress" {
			inProgressProjects++
		}
	}

	c.HTML(http.StatusOK, "manager/profile/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":          "My Profile",
		"ActivePage":         "profile",
		"User":               user,
		"ManagedTeams":       managedTeams,
		"DirectReports":      directReports,
		"AssignedProjects":   len(projects),
		"InProgressProjects": inProgressProjects,
		"ApprovedProjects":   approvedProjects,
		"RejectedProjects":   rejectedProjects,
	}))
}

func (ac *ManagerController) ProjectReviews(c *gin.Context) {
	managerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
		return
	}

	projects, err := ac.ProjectRepo.GetByManagerID(managerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assigned projects")
		return
	}
	approvers, _ := ac.UserRepo.GetUsersByRole("QualityAssurance")

	c.HTML(http.StatusOK, "manager/projects/reviews.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Project Reviews",
		"ActivePage": "project-reviews",
		"projects":   projects,
		"approvers":  approvers,
	}))
}

func (ac *ManagerController) ProjectRequests(c *gin.Context) {
	requests, err := ac.ProjectRepo.GetPendingRequests()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load pending requests")
		return
	}

	teams, err := ac.TeamRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load teams")
		return
	}

	c.HTML(http.StatusOK, "manager/projects/requests.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Project Requests",
		"ActivePage": "project-requests",
		"requests":   requests,
		"teams":      teams,
	}))
}

func (ac *ManagerController) ProcessProjectReview(c *gin.Context) {
	projectID := c.Param("id")
	managerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
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

	if project.ManagerID == nil || *project.ManagerID != managerID {
		c.String(http.StatusForbidden, "You can only review your assigned projects")
		return
	}

	if project.ApprovalStatus != "pending_initial_approval" {
		c.String(http.StatusBadRequest, "Project is not awaiting manager review")
		return
	}

	newProjectStatus := "in_progress"
	if decision == "rejected" {
		newProjectStatus = "cancelled"
	}

	err = ac.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"approval_status": decision,
			"status":          newProjectStatus,
			"updated_at":      time.Now(),
		}).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update project review")
		return
	}
	_ = ac.ProjectRepo.SyncWorkflowStatus(project.ID)

	if project.ApproverID != nil {
		_ = ac.NotificationRepo.CreateForUserWithLink(*project.ApproverID, "Manager "+decision+" project: "+project.Name, "/qa/projects/reviews")
	}
	if project.TeamID != nil {
		var team models.Team
		if teamErr := ac.ProjectRepo.DB.Select("supervisor_id").Where("id = ?", *project.TeamID).First(&team).Error; teamErr == nil {
			_ = ac.NotificationRepo.CreateForUserWithLink(team.SupervisorID, "Manager "+decision+" project: "+project.Name, "/supervisor/projects")
		}
	}

	c.Redirect(http.StatusSeeOther, "/manager/projects/reviews")
}

func (ac *ManagerController) AssignApprover(c *gin.Context) {
	projectID := c.Param("id")
	managerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
		return
	}

	var form struct {
		ApproverID string `form:"approver_id" binding:"required"`
	}
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "Approver is required")
		return
	}

	approverID, err := uuid.Parse(strings.TrimSpace(form.ApproverID))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid approver ID")
		return
	}

	project, err := ac.ProjectRepo.GetByID(projectID)
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}
	if project.ManagerID == nil || *project.ManagerID != managerID {
		c.String(http.StatusForbidden, "You can only assign QA for your own projects")
		return
	}

	qaUsers, err := ac.UserRepo.GetUsersByRole("QualityAssurance")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to validate QA user")
		return
	}
	isQA := false
	for _, user := range qaUsers {
		if user.ID == approverID {
			isQA = true
			break
		}
	}
	if !isQA {
		c.String(http.StatusBadRequest, "Selected user is not a QA approver")
		return
	}

	updates := map[string]interface{}{
		"approver_id": approverID,
		"updated_at":  time.Now(),
	}
	if project.ApprovalStatus == "pending_manager_assignment" {
		updates["approval_status"] = "pending_initial_approval"
	}

	if err := ac.ProjectRepo.DB.Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(updates).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to assign QA approver")
		return
	}

	_ = ac.NotificationRepo.CreateForUserWithLink(approverID, "You were assigned as QA approver for project: "+project.Name, "/qa/projects/reviews")

	c.Redirect(http.StatusSeeOther, "/manager/projects/reviews")
}

func (ac *ManagerController) ApproveProjectRequest(c *gin.Context) {
	managerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid manager ID")
		return
	}

	requestID := c.Param("id")
	teamIDStr := strings.TrimSpace(c.PostForm("team_id"))
	if teamIDStr == "" {
		c.String(http.StatusBadRequest, "Team is required")
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := ac.TeamRepo.GetByID(teamID)
	if err != nil || team == nil || team.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Team not found")
		return
	}

	request, err := ac.ProjectRepo.GetRequestByID(requestID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load request")
		return
	}

	if err := ac.ProjectRepo.UpdateRequestStatus(requestID, "approved"); err != nil {
		c.String(http.StatusInternalServerError, "Failed to approve request")
		return
	}

	project := models.Project{
		ProjectRequestID: &request.ID,
		Name:             request.Title,
		Description:      request.Description,
		ManagerID:        &managerID,
		ApprovalStatus:   "approved",
		Status:           "in_progress",
		StartDate:        time.Now(),
		DueDate:          request.Deadline,
		TeamID:           &team.ID,
	}

	if err := ac.ProjectRepo.Create(&project); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create project")
		return
	}
	if _, err := ac.ProjectRepo.EnsureKickoffTask(&project); err != nil {
		c.String(http.StatusInternalServerError, "Failed to prepare employee taskboard")
		return
	}

	approval := models.Approval{
		ProjectID:  project.ID,
		ApproverID: managerID,
		Stage:      "manager_request_review",
		Status:     "approved",
		Comment:    "Project request approved by manager",
		ApprovedAt: time.Now(),
	}
	_ = ac.ProjectRepo.CreateApproval(&approval)
	_ = ac.ProjectRepo.SyncWorkflowStatus(project.ID)

	_ = ac.NotificationRepo.CreateForUserWithLink(team.SupervisorID, "New project assigned to your team. Assign an employee to start: "+project.Name, "/supervisor/projects")

	c.Redirect(http.StatusSeeOther, "/manager/projects/requests")
}

func (ac *ManagerController) RejectProjectRequest(c *gin.Context) {
	requestID := c.Param("id")
	if err := ac.ProjectRepo.UpdateRequestStatus(requestID, "rejected"); err != nil {
		c.String(http.StatusInternalServerError, "Failed to reject request")
		return
	}

	c.Redirect(http.StatusSeeOther, "/manager/projects/requests")
}

func (ac *ManagerController) Teams(c *gin.Context) {
	teams, err := ac.TeamRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load teams")
		return
	}

	supervisors, _ := ac.TeamRepo.GetSupervisors()
	employees, _ := ac.TeamRepo.GetEmployees()

	c.HTML(http.StatusOK, "manager/teams/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":   "Teams",
		"ActivePage":  "teams",
		"teams":       teams,
		"supervisors": supervisors,
		"employees":   employees,
	}))
}

func (ac *ManagerController) StoreTeam(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	supervisorIDStr := strings.TrimSpace(c.PostForm("supervisor_id"))
	memberIDs := c.PostFormArray("member_ids")
	if len(memberIDs) == 0 {
		memberIDs = c.PostFormArray("member_ids[]")
	}

	if name == "" || supervisorIDStr == "" {
		c.String(http.StatusBadRequest, "Team name and supervisor are required")
		return
	}

	supervisorID, err := uuid.Parse(supervisorIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid supervisor ID")
		return
	}

	team := &models.Team{
		ID:           uuid.New(),
		Name:         name,
		SupervisorID: supervisorID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	team.Members = append(team.Members, models.User{ID: supervisorID})
	for _, memberID := range memberIDs {
		memberUUID, parseErr := uuid.Parse(strings.TrimSpace(memberID))
		if parseErr != nil || memberUUID == supervisorID {
			continue
		}
		team.Members = append(team.Members, models.User{ID: memberUUID})
	}

	if err := ac.TeamRepo.Create(team); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create team")
		return
	}

	c.Redirect(http.StatusSeeOther, "/manager/teams")
}

func (ac *ManagerController) UpdateTeam(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := ac.TeamRepo.GetByID(teamID)
	if err != nil {
		c.String(http.StatusNotFound, "Team not found")
		return
	}

	team.Name = strings.TrimSpace(c.PostForm("name"))
	supervisorIDStr := strings.TrimSpace(c.PostForm("supervisor_id"))
	if team.Name == "" || supervisorIDStr == "" {
		c.String(http.StatusBadRequest, "Team name and supervisor are required")
		return
	}

	supervisorID, err := uuid.Parse(supervisorIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid supervisor ID")
		return
	}

	memberIDs := c.PostFormArray("member_ids")
	if len(memberIDs) == 0 {
		memberIDs = c.PostFormArray("member_ids[]")
	}

	team.SupervisorID = supervisorID
	team.UpdatedAt = time.Now()
	team.Members = []models.User{{ID: supervisorID}}
	for _, memberID := range memberIDs {
		memberUUID, parseErr := uuid.Parse(strings.TrimSpace(memberID))
		if parseErr != nil || memberUUID == supervisorID {
			continue
		}
		team.Members = append(team.Members, models.User{ID: memberUUID})
	}

	if err := ac.TeamRepo.Update(team); err != nil {
		c.String(http.StatusInternalServerError, "Failed to update team")
		return
	}

	c.Redirect(http.StatusSeeOther, "/manager/teams")
}

func (ac *ManagerController) DeleteTeam(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid team ID")
		return
	}

	if err := ac.TeamRepo.Delete(teamID); err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete team")
		return
	}

	c.Redirect(http.StatusSeeOther, "/manager/teams")
}

func (ac *ManagerController) Notifications(c *gin.Context) {
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

	c.HTML(http.StatusOK, "manager/notifications/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":        "Notifications",
		"ActivePage":       "notifications",
		"AllNotifications": notifications,
	}))
}
