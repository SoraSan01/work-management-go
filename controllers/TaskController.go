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

type TaskController struct {
	Repo             *repositories.TaskRepository
	ProjectRepo      *repositories.ProjectRepository
	UserRepo         *repositories.UserRepository
	EventRepo        *repositories.EventRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewTaskController(
	taskRepo *repositories.TaskRepository,
	projectRepo *repositories.ProjectRepository,
	userRepo *repositories.UserRepository,
	eventRepo *repositories.EventRepository,
	notificationRepo *repositories.NotificationRepository,
) *TaskController {
	return &TaskController{
		Repo:             taskRepo,
		ProjectRepo:      projectRepo,
		UserRepo:         userRepo,
		EventRepo:        eventRepo,
		NotificationRepo: notificationRepo,
	}
}

// List tasks
func (tc *TaskController) Index(c *gin.Context) {
	role := c.GetString("role")
	if role == "supervisor" {
		supervisorID, err := uuid.Parse(c.GetString("user_id"))
		if err != nil {
			c.String(http.StatusUnauthorized, "Invalid supervisor ID")
			return
		}

		tasks, err := tc.Repo.GetBySupervisorID(supervisorID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		projects, err := tc.ProjectRepo.GetBySupervisorID(supervisorID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
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

		c.HTML(http.StatusOK, "supervisor/tasks/index.html", utils.TemplateContext(c, gin.H{
			"PageTitle":  "Assign Tasks",
			"ActivePage": "task-assign",
			"Tasks":      tasks,
			"Projects":   projects,
			"Users":      users,
		}))
		return
	}

	tasks, err := tc.Repo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	projects, _ := tc.ProjectRepo.GetAll()
	users, _ := tc.UserRepo.GetAll()
	c.HTML(http.StatusOK, "admin/tasks/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Tasks",
		"ActivePage": "tasks",
		"Tasks":      tasks,
		"Projects":   projects,
		"Users":      users,
	}))
}

// Show create form
func (tc *TaskController) Create(c *gin.Context) {
	projects, err := tc.ProjectRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load projects")
		return
	}

	users, err := tc.UserRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load users")
		return
	}

	c.HTML(http.StatusOK, "admin/tasks/create.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Create Task",
		"ActivePage": "tasks",
		"Projects":   projects,
		"Users":      users,
	}))
}

// Store new task
func (tc *TaskController) Store(c *gin.Context) {
	var input struct {
		ProjectID   string `form:"project_id" binding:"required"`
		Title       string `form:"title" binding:"required"`
		Description string `form:"description"`
		DueDate     string `form:"due_date" binding:"required"`
		Priority    string `form:"priority"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	projectID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid due date")
		return
	}

	project, err := tc.ProjectRepo.GetByID(input.ProjectID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load project")
		return
	}
	if project.ID == uuid.Nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}
	if project.AssignedEmployeeID == nil {
		c.String(http.StatusBadRequest, "Assign an employee to the project before creating tasks")
		return
	}

	assignedTo := *project.AssignedEmployeeID

	role := c.GetString("role")
	if role == "supervisor" {
		supervisorID, err := uuid.Parse(c.GetString("user_id"))
		if err != nil {
			c.String(http.StatusUnauthorized, "Invalid supervisor ID")
			return
		}

		if project.ID == uuid.Nil || project.TeamID == nil {
			c.String(http.StatusForbidden, "You can only assign tasks to projects assigned to your team")
			return
		}

		var teamCount int64
		if err := tc.ProjectRepo.DB.
			Table("teams").
			Where("id = ? AND supervisor_id = ?", *project.TeamID, supervisorID).
			Count(&teamCount).Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to validate team access")
			return
		}
		if teamCount == 0 {
			c.String(http.StatusForbidden, "You can only assign tasks for your own team")
			return
		}

		if assignedTo == supervisorID {
			c.String(http.StatusForbidden, "Supervisors cannot assign project tasks to themselves")
			return
		}

		if project.Status == "completed" || project.Status == "cancelled" {
			c.String(http.StatusBadRequest, "Cannot create new tasks for a completed or final-QA submitted project")
			return
		}

		var memberCount int64
		if err := tc.ProjectRepo.DB.
			Table("team_members").
			Where("team_id = ? AND user_id = ?", *project.TeamID, assignedTo).
			Count(&memberCount).Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to validate assignee")
			return
		}
		if memberCount == 0 {
			c.String(http.StatusForbidden, "The project's assigned employee must belong to your team")
			return
		}
	}

	task := models.Task{
		ProjectID:   projectID,
		AssignedTo:  assignedTo,
		Title:       input.Title,
		Description: input.Description,
		DueDate:     dueDate,
		Priority:    input.Priority,
		Status:      normalizeStatus("todo"),
	}

	if task.Priority == "" {
		task.Priority = "low"
	}

	if err := tc.Repo.Create(&task); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = tc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)

	createdBy := c.GetString("user_name")
	if createdBy == "" {
		createdBy = "A user"
	}
	_ = tc.NotificationRepo.CreateForUserWithLink(task.AssignedTo, createdBy+" assigned you a new task: "+task.Title, "/employee/tasks/board")

	if !task.DueDate.IsZero() {
		start := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 9, 0, 0, 0, task.DueDate.Location())
		end := start.Add(1 * time.Hour)
		event := models.CalendarEvent{
			Title:       "Task Due: " + task.Title,
			Description: "Auto-created from task due date",
			UserID:      &task.AssignedTo,
			StartTime:   start,
			EndTime:     end,
			TaskID:      &task.ID,
			ProjectID:   &task.ProjectID,
		}
		_ = tc.EventRepo.Create(&event)
	}

	redirectPath := "/admin/tasks"
	if role == "supervisor" {
		redirectPath = "/supervisor/tasks"
	}
	c.Redirect(http.StatusFound, redirectPath)
}

// Show edit form
func (tc *TaskController) Edit(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	task, err := tc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	c.HTML(http.StatusOK, "admin/tasks/edit.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Edit Task",
		"ActivePage": "tasks",
		"Task":       task,
	}))
}

// Update task
func (tc *TaskController) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	task, err := tc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	var input struct {
		Title       string `form:"title" binding:"required"`
		Description string `form:"description"`
		Status      string `form:"status" binding:"required"`
		DueDate     string `form:"due_date" binding:"required"`
		Priority    string `form:"priority"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Status = input.Status
	task.DueDate, _ = time.Parse("2006-01-02", input.DueDate)
	task.Priority = input.Priority
	if task.Priority == "" {
		task.Priority = "low"
	}
	task.Status = normalizeStatus(task.Status)

	if err := tc.Repo.Update(task); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = tc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)

	c.Redirect(http.StatusFound, "/admin/tasks")
}

// Delete task
func (tc *TaskController) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	task, err := tc.Repo.GetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	if err := tc.Repo.Delete(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = tc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)
	c.Redirect(http.StatusFound, "/admin/tasks")
}

// StatusColumn represents a Kanban column
type StatusColumn struct {
	Key   string
	Label string
	Color string
}

// Task board view
func (tc *TaskController) Board(c *gin.Context) {
	role := c.GetString("role")
	userIDStr := c.GetString("user_id")

	var tasks []models.Task
	var err error

	if role == "admin" || role == "supervisor" || role == "qualityassurance" || role == "manager" {
		switch role {
		case "supervisor":
			supervisorID, parseErr := uuid.Parse(userIDStr)
			if parseErr != nil {
				c.String(http.StatusUnauthorized, "Invalid supervisor ID")
				return
			}
			tasks, err = tc.Repo.GetBySupervisorID(supervisorID)
		case "qualityassurance":
			qaID, parseErr := uuid.Parse(userIDStr)
			if parseErr != nil {
				c.String(http.StatusUnauthorized, "Invalid quality assurance ID")
				return
			}
			tasks, err = tc.Repo.GetByQATeamMemberID(qaID)
		case "manager":
			managerID, parseErr := uuid.Parse(userIDStr)
			if parseErr != nil {
				c.String(http.StatusUnauthorized, "Invalid manager ID")
				return
			}
			tasks, err = tc.Repo.GetByManagerID(managerID)
		default:
			tasks, err = tc.Repo.GetAll()
		}
	} else {
		userID, _ := uuid.Parse(userIDStr)
		tasks, err = tc.Repo.GetByAssignee(userID)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for i := range tasks {
		tasks[i].Status = normalizeStatus(tasks[i].Status)
	}

	statuses := []StatusColumn{
		{Key: "todo", Label: "To Do", Color: "#3B82F6"},
		{Key: "in_progress", Label: "In Progress", Color: "#F59E0B"},
		{Key: "for_review", Label: "For Review", Color: "#8B5CF6"},
		{Key: "done", Label: "Done", Color: "#10B981"},
	}

	templatePath := "employee/tasks/board.html"
	switch role { // tagged switch on the value of role
	case "admin":
		templatePath = "admin/tasks/board.html"
	case "supervisor":
		templatePath = "supervisor/tasks/board.html"
	case "qualityassurance":
		templatePath = "qa/tasks/board.html"
	case "manager":
		templatePath = "manager/tasks/board.html"
	case "employee":
		templatePath = "employee/tasks/board.html"
	default:
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.HTML(http.StatusOK, templatePath, utils.TemplateContext(c, gin.H{
		"ActivePage": map[string]string{
			"admin":            "task-board",
			"supervisor":       "task-board",
			"qualityassurance": "task-board",
			"manager":          "task-board",
			"employee":         "task-board",
		}[role],
		"Tasks":    tasks,
		"Statuses": statuses,
	}))
}

// add near TaskController
func normalizeStatus(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	switch s {
	case "todo", "to_do":
		return "todo"
	case "in_progress", "inprogress":
		return "in_progress"
	case "for_review", "forreview", "review":
		return "for_review"
	case "done", "completed":
		return "done"
	default:
		return s
	}
}

func (tc *TaskController) StartTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid task id")
		return
	}

	role := c.GetString("role")
	userID, userErr := uuid.Parse(c.GetString("user_id"))
	if userErr != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}

	task, err := tc.Repo.GetByID(taskID)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	if role == "employee" && task.AssignedTo != userID {
		c.String(http.StatusForbidden, "You can only start your own task")
		return
	}

	if role == "supervisor" {
		supervisorID, parseErr := uuid.Parse(c.GetString("user_id"))
		if parseErr != nil {
			c.String(http.StatusUnauthorized, "Invalid supervisor ID")
			return
		}

		var canAccess int64
		if err := tc.Repo.DB.
			Table("tasks").
			Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
			Joins("JOIN teams ON teams.id = projects.team_id").
			Where("tasks.id = ? AND teams.supervisor_id = ?", taskID, supervisorID).
			Count(&canAccess).Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to validate task access")
			return
		}
		if canAccess == 0 {
			c.String(http.StatusForbidden, "You can only start tasks assigned to your own team")
			return
		}
	}

	if err := tc.Repo.StartTask(taskID); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if task, err := tc.Repo.GetByID(taskID); err == nil {
		_ = tc.ProjectRepo.SyncWorkflowStatus(task.ProjectID)
		_ = tc.NotificationRepo.CreateForUserWithLink(task.AssignedTo, "Task started: "+task.Title, "/employee/tasks/board")
	}

	templatePath := "/employee/tasks/board"

	switch role { // tagged switch on the value of role
	case "admin":
		templatePath = "/admin/tasks/board"
	case "supervisor":
		templatePath = "/supervisor/tasks/board"
	case "employee":
		templatePath = "/employee/tasks/board"
	default:
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Redirect(http.StatusSeeOther, templatePath)
}

func (tc *TaskController) PauseTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := tc.Repo.GetByID(taskID)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}
	role := c.GetString("role")
	if role != "admin" && task.AssignedTo != userID {
		c.String(http.StatusForbidden, "You can only pause your own task")
		return
	}
	if task.StartedAt == nil {
		c.String(http.StatusBadRequest, "Task is not currently running")
		return
	}

	elapsed := int64(time.Since(*task.StartedAt).Seconds())
	if elapsed > 0 {
		if err := tc.Repo.AddElapsedDuration(taskID, elapsed); err != nil {
			c.String(http.StatusInternalServerError, "Failed to pause task")
			return
		}
	}

	if err := tc.Repo.PauseTask(taskID); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (tc *TaskController) ResumeTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := tc.Repo.GetByID(taskID)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}
	role := c.GetString("role")
	if role != "admin" && task.AssignedTo != userID {
		c.String(http.StatusForbidden, "You can only resume your own task")
		return
	}

	if err := tc.Repo.ResumeTask(taskID); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
