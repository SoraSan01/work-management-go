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

type ProjectController struct {
	Repo      *repositories.ProjectRepository
	UserRepo  *repositories.UserRepository // Add this
	TeamRepo  *repositories.TeamRepository
	EventRepo *repositories.EventRepository
}

func NewProjectController(repo *repositories.ProjectRepository, userRepo *repositories.UserRepository, teamRepo *repositories.TeamRepository, eventRepo *repositories.EventRepository) *ProjectController {
	return &ProjectController{
		Repo:      repo,
		UserRepo:  userRepo,
		TeamRepo:  teamRepo,
		EventRepo: eventRepo,
	}
}

// List projects
func (pc *ProjectController) Index(c *gin.Context) {
	projects, err := pc.Repo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	managers, _ := pc.UserRepo.GetUsersByRole("Manager")
	approvers, _ := pc.UserRepo.GetUsersByRole("QualityAssurance")
	c.HTML(http.StatusOK, "admin/projects/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle": "Projects",
		"ActivePage": "projects",
		"projects":  projects,
		"managers":  managers,
		"approvers": approvers,
	}))
}

// Show create form
func (pc *ProjectController) Create(c *gin.Context) {
	// Fetch managers and approvers
	managers, err := pc.UserRepo.GetUsersByRole("Manager")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	approvers, err := pc.UserRepo.GetUsersByRole("QualityAssurance")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "admin/projects/create.html", utils.TemplateContext(c, gin.H{
		"title":     "Create Project",
		"ActivePage": "projects",
		"managers":  managers,
		"approvers": approvers,
	}))
}

// Store project
func (pc *ProjectController) Store(c *gin.Context) {
	// Bind form data
	var form struct {
		Name        string `form:"Name" binding:"required"`
		Description string `form:"Description"`
		ManagerID   string `form:"ManagerID" binding:"required"`
		ApproverID  string `form:"ApproverID" binding:"required"`
		StartDate   string `form:"StartDate"`
		DueDate     string `form:"DueDate"`
		Status      string `form:"Status"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Convert ManagerID to uuid.UUID
	managerID, err := uuid.Parse(form.ManagerID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Manager ID")
		return
	}

	// Convert ApproverID to uuid.UUID
	approverID, err := uuid.Parse(form.ApproverID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Approver ID")
		return
	}

	project := models.Project{
		Name:        form.Name,
		Description: form.Description,
		ManagerID:   &managerID,
		ApproverID:  &approverID,
		Status:      form.Status,
	}

	// Optional: parse dates if your model uses time.Time
	if form.StartDate != "" {
		start, err := time.Parse("2006-01-02", form.StartDate)
		if err == nil {
			project.StartDate = start
		}
	}
	if form.DueDate != "" {
		due, err := time.Parse("2006-01-02", form.DueDate)
		if err == nil {
			project.DueDate = due
		}
	}

	if err := pc.Repo.Create(&project); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if !project.DueDate.IsZero() {
		start := time.Date(project.DueDate.Year(), project.DueDate.Month(), project.DueDate.Day(), 9, 0, 0, 0, project.DueDate.Location())
		end := start.Add(1 * time.Hour)
		event := models.CalendarEvent{
			Title:       "Project Due: " + project.Name,
			Description: "Auto-created from project due date",
			UserID:      project.ManagerID,
			StartTime:   start,
			EndTime:     end,
			ProjectID:   &project.ID,
		}
		_ = pc.EventRepo.Create(&event)
	}

	c.Redirect(http.StatusFound, "/admin/projects")
}

// Show edit form
func (pc *ProjectController) Edit(c *gin.Context) {
	id := c.Param("id")
	project, err := pc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// Fetch managers and approvers
	managers, err := pc.UserRepo.GetUsersByRole("Manager")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	approvers, err := pc.UserRepo.GetUsersByRole("QualityAssurance")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "admin/projects/edit.html", utils.TemplateContext(c, gin.H{
		"title":     "Edit Project",
		"ActivePage": "projects",
		"project":   project,
		"managers":  managers,
		"approvers": approvers,
	}))
}

// Update project
func (pc *ProjectController) Update(c *gin.Context) {
	id := c.Param("id")
	project, err := pc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Bind form data with proper struct
	var form struct {
		Name        string `form:"Name"`
		Description string `form:"Description"`
		ManagerID   string `form:"ManagerID"`
		ApproverID  string `form:"ApproverID"`
		StartDate   string `form:"StartDate"`
		DueDate     string `form:"DueDate"`
		Status      string `form:"Status"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "Form binding error: "+err.Error())
		return
	}

	// Update project fields
	project.Name = form.Name
	project.Description = form.Description
	project.Status = form.Status

	// Parse and update ManagerID if provided
	if form.ManagerID != "" {
		managerID, err := uuid.Parse(strings.TrimSpace(form.ManagerID))
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Manager ID")
			return
		}
		project.ManagerID = &managerID
	}

	// Parse and update ApproverID if provided
	if form.ApproverID != "" {
		approverID, err := uuid.Parse(strings.TrimSpace(form.ApproverID))
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Approver ID")
			return
		}
		project.ApproverID = &approverID
	}

	// Parse dates if provided
	if form.StartDate != "" {
		start, err := time.Parse("2006-01-02", form.StartDate)
		if err == nil {
			project.StartDate = start
		}
	}
	if form.DueDate != "" {
		due, err := time.Parse("2006-01-02", form.DueDate)
		if err == nil {
			project.DueDate = due
		}
	}

	if err := pc.Repo.Update(&project); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = pc.Repo.SyncWorkflowStatus(project.ID)

	c.Redirect(http.StatusFound, "/admin/projects")
}

// Delete project
func (pc *ProjectController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := pc.Repo.Delete(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/admin/projects")
}

// List pending project requests
func (pc *ProjectController) Requests(c *gin.Context) {
	requests, err := pc.Repo.GetPendingRequests()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	teams, _ := pc.TeamRepo.GetAll()

	// Capitalize status before sending to template
	for i := range requests {
		requests[i].Status = strings.Title(requests[i].Status)
	}

	c.HTML(http.StatusOK, "admin/projects/requests.html", utils.TemplateContext(c, gin.H{
		"title":    "Pending Requests",
		"PageTitle": "Project Requests",
		"ActivePage": "project-requests",
		"requests": requests,
		"teams":    teams,
	}))
}

// Approve a project request
func (pc *ProjectController) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	teamIDStr := c.PostForm("team_id")

	// Update request status
	err := pc.Repo.UpdateRequestStatus(id, "approved")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Get the request with all related data
	request, err := pc.Repo.GetRequestByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Determine manager/team
	var defaultManager models.User
	var assignedTeamID *uuid.UUID

	if teamIDStr != "" {
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid team ID")
			return
		}
		team, err := pc.TeamRepo.GetByID(teamID)
		if err != nil {
			c.String(http.StatusInternalServerError, "Team not found")
			return
		}
		assignedTeamID = &team.ID
		defaultManager = team.Supervisor
	}

	if defaultManager.ID == uuid.Nil {
		// Fallback: get first available manager
		result := pc.UserRepo.DB.Where("role_id = (SELECT id FROM roles WHERE name = 'Manager') AND is_active = true").
			First(&defaultManager)
		if result.Error != nil {
			c.String(http.StatusInternalServerError, "No available manager found")
			return
		}
	}

	// Create project with pending approval status
	project := models.Project{
		ProjectRequestID: &request.ID,
		Name:             request.Title,
		Description:      request.Description,
		ManagerID:        &defaultManager.ID,
		ApproverID:       nil, // Will be assigned later
		ApprovalStatus:   "pending_initial_approval",
		Status:           "draft",
		StartDate:        time.Now(),
		DueDate:          request.Deadline,
		TeamID:           assignedTeamID,
	}

	if err := pc.Repo.Create(&project); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = pc.Repo.SyncWorkflowStatus(project.ID)

	// Create initial approval record for manager
	approval := models.Approval{
		ProjectID:  project.ID,
		ApproverID: defaultManager.ID,
		Stage:      "manager_assignment",
		Status:     "approved", // Auto-approved since we assigned them
		Comment:    "Manager assigned to project",
		ApprovedAt: time.Now(),
	}

	_ = pc.Repo.CreateApproval(&approval)

	// Redirect to project approval page instead of requests list
	c.Redirect(http.StatusSeeOther, "/admin/projects/"+project.ID.String()+"/approval")
}

// Reject a project request
func (pc *ProjectController) RejectRequest(c *gin.Context) {
	id := c.Param("id")

	var form struct {
		Reason string `form:"reason"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Update request status with reason
	err := pc.Repo.DB.Model(&models.ProjectRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           "rejected",
			"rejection_reason": form.Reason,
		}).Error

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/projects/requests")
}

func (pc *ProjectController) ShowApproval(c *gin.Context) {
	id := c.Param("id")
	project, err := pc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	approvals, err := pc.Repo.GetProjectApprovals(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	approvers, _ := pc.UserRepo.GetUsersByRole("QualityAssurance")

	c.HTML(http.StatusOK, "admin/projects/approval.html", utils.TemplateContext(c, gin.H{
		"title":     "Project Approval",
		"PageTitle": "Project Approval",
		"ActivePage": "projects",
		"project":   project,
		"approvers": approvers,
		"approvals": approvals,
	}))
}

func (pc *ProjectController) ProcessApproval(c *gin.Context) {
	projectID := c.Param("id")

	var form struct {
		Status     string `form:"status" binding:"required"`
		Comment    string `form:"comment"`
		ApprovalID string `form:"approval_id"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Get current user (you'll need to implement this from your auth middleware)
	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		c.String(http.StatusUnauthorized, "User not authenticated")
		return
	}

	// If this is for an existing approval record
	if form.ApprovalID != "" {
		err := pc.Repo.UpdateApproval(form.ApprovalID, form.Status, form.Comment)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		// Create new approval record
		approverID, err := uuid.Parse(currentUserID)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid user ID")
			return
		}

		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid project ID")
			return
		}

		approval := models.Approval{
			ProjectID:  projectUUID,
			ApproverID: approverID,
			Stage:      "initial",
			Status:     form.Status,
			Comment:    form.Comment,
		}

		if err := pc.Repo.CreateApproval(&approval); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update project approval status based on decision
	switch form.Status {
	case "approved":
		if err := pc.Repo.UpdateProjectApprovalStatus(projectID, "approved"); err != nil {
			c.String(http.StatusInternalServerError, "Failed to update project approval status")
			return
		}
		// Also update project status to active
		if err := pc.Repo.DB.Model(&models.Project{}).Where("id = ?", projectID).
			Update("status", "in_progress").Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to update project status")
			return
		}
	case "rejected":
		if err := pc.Repo.UpdateProjectApprovalStatus(projectID, "rejected"); err != nil {
			c.String(http.StatusInternalServerError, "Failed to update project approval status")
			return
		}
	}
	if projectUUID, parseErr := uuid.Parse(projectID); parseErr == nil {
		_ = pc.Repo.SyncWorkflowStatus(projectUUID)
	}

	c.Redirect(http.StatusSeeOther, "/admin/projects")
}

func (pc *ProjectController) AssignApprover(c *gin.Context) {
	projectID := c.Param("id")

	var form struct {
		ApproverID string `form:"approver_id" binding:"required"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	approverID, err := uuid.Parse(form.ApproverID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid approver ID")
		return
	}

	// Update project with approver
	err = pc.Repo.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		Updates(map[string]interface{}{
			"approver_id":     &approverID,
			"approval_status": "pending_initial_approval",
		}).Error

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Create approval record
	projectUUID, _ := uuid.Parse(projectID)
	approval := models.Approval{
		ProjectID:  projectUUID,
		ApproverID: approverID,
		Stage:      "initial",
		Status:     "pending",
		Comment:    "Waiting for initial approval",
	}

	if err := pc.Repo.CreateApproval(&approval); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = pc.Repo.SyncWorkflowStatus(projectUUID)

	c.Redirect(http.StatusSeeOther, "/admin/projects/"+projectID+"/approval")
}
