package controllers

import (
	"net/http"
	"strings"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskReviewController struct {
	TaskRepo    *repositories.TaskRepository
	ProjectRepo *repositories.ProjectRepository
	CommentRepo *repositories.TaskReviewCommentRepository
	NotifyRepo  *repositories.NotificationRepository
}

func NewTaskReviewController(taskRepo *repositories.TaskRepository, projectRepo *repositories.ProjectRepository, commentRepo *repositories.TaskReviewCommentRepository, notifyRepo *repositories.NotificationRepository) *TaskReviewController {
	return &TaskReviewController{
		TaskRepo:    taskRepo,
		ProjectRepo: projectRepo,
		CommentRepo: commentRepo,
		NotifyRepo:  notifyRepo,
	}
}

func (trc *TaskReviewController) ListComments(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := trc.TaskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	role := c.GetString("role")
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	if role == "qualityassurance" {
		canReview, checkErr := trc.TaskRepo.CanQAReviewTask(taskID, userID)
		if checkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate task review access"})
			return
		}
		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only view comments for tasks in your assigned teams"})
			return
		}
	}
	if role == "manager" {
		canReview, checkErr := trc.TaskRepo.CanManagerReviewTask(taskID, userID)
		if checkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate task review access"})
			return
		}
		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only view comments for tasks in your assigned projects"})
			return
		}
	}

	// Assigned employee can read comments on their task; admin can read all.
	if role != "admin" && role != "qualityassurance" && role != "manager" && task.AssignedTo != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot view comments for this task"})
		return
	}

	comments, err := trc.CommentRepo.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
		return
	}

	out := make([]gin.H, 0, len(comments))
	for _, item := range comments {
		out = append(out, gin.H{
			"id":           item.ID,
			"task_id":      item.TaskID,
			"action":       item.Action,
			"message":      item.Message,
			"created_at":   item.CreatedAt,
			"creator_id":   item.CreatedBy,
			"creator_name": item.Creator.Name(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"comments": out,
	})
}

func (trc *TaskReviewController) SubmitReview(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := trc.TaskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	role := c.GetString("role")
	if role != "admin" && role != "qualityassurance" && role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin, quality assurance, or manager can review tasks"})
		return
	}

	var input struct {
		Message string `json:"message"`
		Action  string `json:"action" binding:"required"` // approved, rejected, commented
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action == "approve" {
		action = "approved"
	}
	if action == "reject" {
		action = "rejected"
	}
	if action == "" {
		action = "commented"
	}
	if action != "approved" && action != "rejected" && action != "commented" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be approved, rejected, or commented"})
		return
	}

	reviewerID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	if role == "qualityassurance" {
		canReview, checkErr := trc.TaskRepo.CanQAReviewTask(taskID, reviewerID)
		if checkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate task review access"})
			return
		}
		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only review tasks for teams where you are assigned"})
			return
		}
	}
	if role == "manager" {
		canReview, checkErr := trc.TaskRepo.CanManagerReviewTask(taskID, reviewerID)
		if checkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate task review access"})
			return
		}
		if !canReview {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only review tasks in your assigned projects"})
			return
		}
	}

	comment := models.TaskReviewComment{
		TaskID:    taskID,
		CreatedBy: reviewerID,
		Message:   strings.TrimSpace(input.Message),
		Action:    action,
	}
	if comment.Message == "" {
		switch action {
		case "approved":
			comment.Message = "Task approved."
		case "rejected":
			comment.Message = "Task rejected. Please revise and resubmit."
		default:
			comment.Message = "Task review comment."
		}
	}

	if err := trc.CommentRepo.Create(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save review comment"})
		return
	}

	if action == "approved" {
		if task.StartedAt != nil {
			elapsed := int64(time.Since(*task.StartedAt).Seconds())
			if elapsed > 0 {
				_ = trc.TaskRepo.AddElapsedDuration(task.ID, elapsed)
			}
		}
		task.Status = "done"
		task.StartedAt = nil
		if err := trc.TaskRepo.Update(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
			return
		}
		_ = trc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)
	}

	if action == "rejected" {
		if task.StartedAt != nil {
			elapsed := int64(time.Since(*task.StartedAt).Seconds())
			if elapsed > 0 {
				_ = trc.TaskRepo.AddElapsedDuration(task.ID, elapsed)
			}
		}
		task.Status = "in_progress"
		task.StartedAt = nil
		if err := trc.TaskRepo.Update(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
			return
		}
		_ = trc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)
	}

	_ = trc.NotifyRepo.CreateForUserWithLink(task.AssignedTo, "Task review update: "+task.Title+" ("+action+")", "/employee/tasks/board")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": comment,
		"status":  task.Status,
	})
}
