package repositories

import (
	"work-management-system/models"

	"gorm.io/gorm"
)

type ReportRepository struct {
	DB *gorm.DB
}

type EmployeeRoleStat struct {
	RoleName string
	Count    int64
}

type EmployeeDepartmentStat struct {
	DepartmentName string
	Count          int64
}

type EmployeeReportData struct {
	TotalEmployees    int64
	ActiveEmployees   int64
	InactiveEmployees int64
	RoleStats         []EmployeeRoleStat
	DepartmentStats   []EmployeeDepartmentStat
	RecentEmployees   []models.User
}

type StatusCount struct {
	Status string
	Count  int64
}

type PriorityCount struct {
	Priority string
	Count    int64
}

type ProjectReportData struct {
	TotalProjects        int64
	CompletedProjects    int64
	InProgressProjects   int64
	PendingApprovalCount int64
	StatusBreakdown      []StatusCount
	RecentProjects       []models.Project
}

type TaskReportData struct {
	TotalTasks      int64
	DoneTasks       int64
	ForReviewTasks  int64
	InProgressTasks int64
	StatusBreakdown []StatusCount
	PriorityStats   []PriorityCount
	RecentTasks     []models.Task
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{DB: db}
}

func (r *ReportRepository) EmployeeReport() (EmployeeReportData, error) {
	result := EmployeeReportData{}

	if err := r.DB.Model(&models.User{}).Count(&result.TotalEmployees).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&result.ActiveEmployees).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.User{}).Where("is_active = ?", false).Count(&result.InactiveEmployees).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Table("roles").
		Select("roles.name as role_name, COUNT(users.id) as count").
		Joins("LEFT JOIN users ON users.role_id = roles.id").
		Group("roles.name").
		Order("count DESC, roles.name ASC").
		Scan(&result.RoleStats).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Table("departments").
		Select("departments.name as department_name, COUNT(users.id) as count").
		Joins("LEFT JOIN users ON users.department_id = departments.id").
		Group("departments.name").
		Order("count DESC, departments.name ASC").
		Scan(&result.DepartmentStats).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Preload("Role").
		Preload("Department").
		Order("created_at DESC").
		Limit(10).
		Find(&result.RecentEmployees).Error; err != nil {
		return result, err
	}

	return result, nil
}

func (r *ReportRepository) ProjectReport() (ProjectReportData, error) {
	result := ProjectReportData{}

	if err := r.DB.Model(&models.Project{}).Count(&result.TotalProjects).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Project{}).Where("status = ?", "completed").Count(&result.CompletedProjects).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Project{}).Where("status = ?", "in_progress").Count(&result.InProgressProjects).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Project{}).Where("approval_status LIKE ?", "pending%").Count(&result.PendingApprovalCount).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Model(&models.Project{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Order("count DESC, status ASC").
		Scan(&result.StatusBreakdown).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Preload("Manager").
		Preload("Approver").
		Preload("Team").
		Order("created_at DESC").
		Limit(10).
		Find(&result.RecentProjects).Error; err != nil {
		return result, err
	}

	return result, nil
}

func (r *ReportRepository) TaskReport() (TaskReportData, error) {
	result := TaskReportData{}

	if err := r.DB.Model(&models.Task{}).Count(&result.TotalTasks).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Task{}).Where("status = ?", "done").Count(&result.DoneTasks).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Task{}).Where("status = ?", "for_review").Count(&result.ForReviewTasks).Error; err != nil {
		return result, err
	}
	if err := r.DB.Model(&models.Task{}).Where("status = ?", "in_progress").Count(&result.InProgressTasks).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Model(&models.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Order("count DESC, status ASC").
		Scan(&result.StatusBreakdown).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Model(&models.Task{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Order("count DESC, priority ASC").
		Scan(&result.PriorityStats).Error; err != nil {
		return result, err
	}

	if err := r.DB.
		Preload("Project").
		Preload("Assignee").
		Order("created_at DESC").
		Limit(10).
		Find(&result.RecentTasks).Error; err != nil {
		return result, err
	}

	return result, nil
}
