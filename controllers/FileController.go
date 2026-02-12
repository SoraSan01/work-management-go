package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileController struct {
	Repo        *repositories.FileRepository
	TaskRepo    *repositories.TaskRepository
	ProjectRepo *repositories.ProjectRepository
}

func NewFileController(repo *repositories.FileRepository, taskRepo *repositories.TaskRepository, projectRepo *repositories.ProjectRepository) *FileController {
	return &FileController{Repo: repo, TaskRepo: taskRepo, ProjectRepo: projectRepo}
}

func (fc *FileController) Upload(c *gin.Context) {
	role := c.GetString("role")
	if role == "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "supervisors cannot upload task output files"})
		return
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	userID := c.GetString("user_id")
	uploaderID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	// Ownership check
	task, err := fc.TaskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if task.AssignedTo != uploaderID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot upload to this task"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Size limit (e.g., 20MB)
	if file.Size > 20<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 20MB)"})
		return
	}

	// Create task-specific upload folder
	uploadDir := filepath.Join("uploads", "tasks", taskID.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload folder"})
		return
	}

	// Save file with unique name (UUID)
	filename := uuid.New().String() + "_" + filepath.Base(file.Filename)
	path := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// File type categorization (optional)
	fileType := "other"
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".pdf" || ext == ".docx" {
		fileType = "output"
	}

	record := models.File{
		ProjectID:  task.ProjectID,
		TaskID:     taskID,
		UploadedBy: uploaderID,
		FilePath:   path,
		FileType:   fileType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := fc.Repo.Create(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to save file record",
			"details": err.Error(),
		})
		return
	}

	if task.StartedAt != nil {
		elapsed := int64(time.Since(*task.StartedAt).Seconds())
		if elapsed > 0 {
			_ = fc.TaskRepo.AddElapsedDuration(taskID, elapsed)
		}
	}

	if err := fc.TaskRepo.MarkForReviewAfterUpload(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update task status",
			"details": err.Error(),
		})
		return
	}
	_ = fc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"file":      record,
		"file_name": extractDisplayFileName(record.FilePath),
	})
}

func (fc *FileController) ListByTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := fc.TaskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	role := c.GetString("role")
	userID := c.GetString("user_id")
	requesterID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	if role != "admin" && role != "supervisor" && role != "qualityassurance" && task.AssignedTo != requesterID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot view files for this task"})
		return
	}

	files, err := fc.Repo.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load files"})
		return
	}

	out := make([]gin.H, 0, len(files))
	for _, file := range files {
		out = append(out, gin.H{
			"id":           file.ID,
			"file_path":    file.FilePath,
			"download_url": buildPublicFileURL(file.FilePath),
			"file_name":    extractDisplayFileName(file.FilePath),
			"uploaded_at":  file.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"files":   out,
	})
}

func (fc *FileController) ListCustomerRequestFiles(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("requestId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	customerID, err := uuid.Parse(c.GetString("customer_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid customer"})
		return
	}

	var files []models.File
	err = fc.Repo.DB.
		Table("files").
		Select("files.*").
		Joins("JOIN tasks ON tasks.id = files.task_id").
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("projects.project_request_id = ? AND project_requests.customer_id = ? AND tasks.status = ?", requestID, customerID, "done").
		Order("files.created_at DESC").
		Scan(&files).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load project files"})
		return
	}

	out := make([]gin.H, 0, len(files))
	for _, file := range files {
		out = append(out, gin.H{
			"id":           file.ID,
			"file_path":    file.FilePath,
			"file_name":    extractDisplayFileName(file.FilePath),
			"download_url": buildPublicFileURL(file.FilePath),
			"uploaded_at":  file.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"files":   out,
	})
}

func (fc *FileController) UploadProjectFinal(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	supervisorID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid supervisor")
		return
	}

	project, err := fc.ProjectRepo.GetByID(projectID.String())
	if err != nil || project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}
	if project.TeamID == nil {
		c.String(http.StatusForbidden, "Project is not assigned to a team")
		return
	}

	var canAccess int64
	if err := fc.ProjectRepo.DB.
		Table("teams").
		Where("id = ? AND supervisor_id = ?", *project.TeamID, supervisorID).
		Count(&canAccess).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to validate project access")
		return
	}
	if canAccess == 0 {
		c.String(http.StatusForbidden, "You can only upload final files for your own team projects")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Final file is required")
		return
	}
	if file.Size > 50<<20 {
		c.String(http.StatusBadRequest, "File too large (max 50MB)")
		return
	}

	uploadDir := filepath.Join("uploads", "projects", projectID.String(), "final")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create upload folder")
		return
	}

	filename := uuid.New().String() + "_" + filepath.Base(file.Filename)
	path := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file")
		return
	}

	record := models.ProjectFinalFile{
		ProjectID:  projectID,
		UploadedBy: supervisorID,
		FilePath:   path,
		FileName:   file.Filename,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := fc.Repo.DB.Create(&record).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file record")
		return
	}

	if project.ApproverID != nil {
		_ = fc.ProjectRepo.DB.Model(&models.Project{}).
			Where("id = ?", projectID).
			Update("updated_at", time.Now()).Error
	}

	c.Redirect(http.StatusSeeOther, "/supervisor/projects")
}

func extractDisplayFileName(filePath string) string {
	base := filepath.Base(filePath)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return base
}

func buildPublicFileURL(filePath string) string {
	clean := strings.ReplaceAll(filePath, "\\", "/")
	if strings.HasPrefix(clean, "/") {
		return clean
	}
	return "/" + clean
}
