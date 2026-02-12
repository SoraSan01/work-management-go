package repositories

import (
	"errors"
	"strings"
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	DB *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{DB: db}
}

// Get all projects
func (r *ProjectRepository) GetAll() ([]models.Project, error) {
	var projects []models.Project
	err := r.DB.
		Preload("Manager").
		Preload("Approver").
		Preload("Tasks").
		Preload("Team").
		Preload("Team.Supervisor").
		Preload("Team.Members").
		Find(&projects).Error

	if err != nil {
		return projects, err
	}

	r.hydrateMissingManagers(projects)

	return projects, err
}

// Get project by ID
func (r *ProjectRepository) GetByID(id string) (models.Project, error) {
	var project models.Project
	err := r.DB.
		Preload("Manager").
		Preload("Approver").
		Preload("Tasks").
		Preload("Team").
		Preload("Team.Supervisor").
		Preload("Team.Members").
		Preload("FinalQASubmitter").
		Preload("FinalQAReviewer").
		Preload("FinalQASender").
		First(&project, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return project, nil
	}
	return project, err
}

// Create project
func (r *ProjectRepository) Create(project *models.Project) error {
	return r.DB.Create(project).Error
}

// Update project
func (r *ProjectRepository) Update(project *models.Project) error {
	return r.DB.Save(project).Error
}

// Delete project
func (r *ProjectRepository) Delete(id string) error {
	return r.DB.Delete(&models.Project{}, "id = ?", id).Error
}

// Get pending project requests
func (r *ProjectRepository) GetPendingRequests() ([]models.ProjectRequest, error) {
	var requests []models.ProjectRequest
	// Preload Customer and nested User
	err := r.DB.Preload("Customer.User").Where("status = ?", "pending").Find(&requests).Error
	return requests, err
}

// Also update GetRequestByID
func (r *ProjectRepository) GetRequestByID(id string) (models.ProjectRequest, error) {
	var request models.ProjectRequest
	err := r.DB.Preload("Customer.User").First(&request, "id = ?", id).Error
	return request, err
}

// Update status of a project request
func (r *ProjectRepository) UpdateRequestStatus(id string, status string) error {
	return r.DB.Model(&models.ProjectRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ProjectRepository) CreateApproval(approval *models.Approval) error {
	return r.DB.Create(approval).Error
}

func (r *ProjectRepository) GetProjectApprovals(projectID string) ([]models.Approval, error) {
	var approvals []models.Approval
	err := r.DB.Preload("Approver").Where("project_id = ?", projectID).Order("created_at DESC").Find(&approvals).Error
	return approvals, err
}

func (r *ProjectRepository) UpdateProjectApprovalStatus(projectID string, status string) error {
	return r.DB.Model(&models.Project{}).Where("id = ?", projectID).Update("approval_status", status).Error
}

func (r *ProjectRepository) UpdateApproval(id string, status string, comment string) error {
	updates := map[string]interface{}{
		"status":  status,
		"comment": comment,
	}

	switch status {
	case "approved":
		updates["approved_at"] = time.Now()
	case "rejected":
		updates["rejected_at"] = time.Now()
	}

	return r.DB.Model(&models.Approval{}).Where("id = ?", id).Updates(updates).Error
}

// Get projects by approval status
func (r *ProjectRepository) GetProjectsByApprovalStatus(status string) ([]models.Project, error) {
	var projects []models.Project
	err := r.DB.Preload("Manager").Preload("Approver").Preload("Approvals").
		Where("approval_status = ?", status).Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) GetBySupervisorID(supervisorID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	err := r.DB.
		Joins("JOIN teams ON teams.id = projects.team_id").
		Where("teams.supervisor_id = ?", supervisorID).
		Preload("Manager").
		Preload("Approver").
		Preload("Tasks").
		Preload("Team").
		Preload("Team.Supervisor").
		Preload("Team.Members").
		Preload("FinalQASubmitter").
		Preload("FinalQAReviewer").
		Preload("FinalQASender").
		Find(&projects).Error
	if err == nil {
		r.hydrateMissingManagers(projects)
	}
	return projects, err
}

func (r *ProjectRepository) GetByAssignedQAID(qaID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	err := r.DB.
		Where("approver_id = ?", qaID).
		Preload("Manager").
		Preload("Approver").
		Preload("Tasks").
		Preload("Team").
		Preload("Team.Members").
		Preload("Team.Supervisor").
		Preload("FinalQASubmitter").
		Preload("FinalQAReviewer").
		Preload("FinalQASender").
		Order("updated_at DESC").
		Find(&projects).Error
	if err == nil {
		r.hydrateMissingManagers(projects)
	}
	return projects, err
}

func (r *ProjectRepository) GetByManagerID(managerID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	err := r.DB.
		Where("manager_id = ?", managerID).
		Preload("Manager").
		Preload("Approver").
		Preload("Tasks").
		Preload("Team").
		Preload("Team.Members").
		Preload("Team.Supervisor").
		Preload("FinalQASubmitter").
		Preload("FinalQAReviewer").
		Preload("FinalQASender").
		Order("updated_at DESC").
		Find(&projects).Error
	if err == nil {
		r.hydrateMissingManagers(projects)
	}
	return projects, err
}

func (r *ProjectRepository) hydrateMissingManagers(projects []models.Project) {
	var managerIDs []uuid.UUID
	managerIndices := make(map[uuid.UUID][]int)

	for i, p := range projects {
		if p.ManagerID != nil && p.Manager == nil {
			managerIDs = append(managerIDs, *p.ManagerID)
			managerIndices[*p.ManagerID] = append(managerIndices[*p.ManagerID], i)
		}
	}
	if len(managerIDs) == 0 {
		return
	}

	var managers []models.User
	if err := r.DB.Where("id IN ?", managerIDs).Find(&managers).Error; err != nil {
		return
	}
	for _, manager := range managers {
		for _, idx := range managerIndices[manager.ID] {
			managerCopy := manager
			projects[idx].Manager = &managerCopy
		}
	}
}

func (r *ProjectRepository) SyncWorkflowStatus(projectID uuid.UUID) error {
	var project models.Project
	if err := r.DB.Select("id", "manager_id", "approval_status", "final_qa_status", "final_qa_sent", "status").
		First(&project, "id = ?", projectID).Error; err != nil {
		return err
	}

	var totalTasks int64
	if err := r.DB.Model(&models.Task{}).Where("project_id = ?", projectID).Count(&totalTasks).Error; err != nil {
		return err
	}

	var doneTasks int64
	if err := r.DB.Model(&models.Task{}).
		Where("project_id = ? AND status IN ?", projectID, []string{"done", "completed"}).
		Count(&doneTasks).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	switch project.FinalQAStatus {
	case "approved":
		if project.FinalQASent {
			updates["status"] = "completed"
		} else {
			updates["status"] = "in_progress"
		}
	case "pending_qa_review":
		updates["status"] = "completed"
	case "rejected":
		updates["status"] = "in_progress"
	default:
		switch project.ApprovalStatus {
		case "rejected":
			updates["status"] = "cancelled"
		case "approved":
			if totalTasks > 0 && doneTasks == totalTasks {
				updates["status"] = "completed"
			} else if project.Status != "draft" {
				updates["status"] = "in_progress"
			}
		}
	}

	nextStatus := project.Status
	if v, ok := updates["status"].(string); ok {
		nextStatus = v
	}
	if nextStatus == "completed" && strings.HasPrefix(project.ApprovalStatus, "pending") {
		updates["approval_status"] = "approved"
	}
	if project.ApprovalStatus == "pending_manager_assignment" && project.ManagerID != nil {
		updates["approval_status"] = "pending_initial_approval"
	}

	if len(updates) == 1 {
		return nil
	}

	return r.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		Updates(updates).Error
}
