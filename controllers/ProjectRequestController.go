package controllers

import (
	"net/http"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectRequestController struct {
	Repo *repositories.ProjectRequestRepository
}

func NewProjectRequestController(repo *repositories.ProjectRequestRepository) *ProjectRequestController {
	return &ProjectRequestController{Repo: repo}
}

// GET /customer/projects
func (prc *ProjectRequestController) Index(c *gin.Context) {
	// Get the customer ID from context
	customerIDVal, exists := c.Get("customer_id")
	if !exists {
		c.String(http.StatusUnauthorized, "Customer not logged in")
		return
	}

	customerIDStr, ok := customerIDVal.(string)
	if !ok || customerIDStr == "" {
		c.String(http.StatusUnauthorized, "Invalid customer ID")
		return
	}

	// Parse UUID safely
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid customer ID format")
		return
	}

	// Fetch projects for this customer
	projects, err := prc.Repo.GetAllByCustomerID(customerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching projects")
		return
	}

	c.HTML(http.StatusOK, "customer/projects/index.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Project Requests",
		"ActivePage": "project-requests",
		"projects":   projects,
	}))
}

// GET /customer/projects/done
func (prc *ProjectRequestController) DoneProjects(c *gin.Context) {
	customerIDVal, exists := c.Get("customer_id")
	if !exists {
		c.String(http.StatusUnauthorized, "Customer not logged in")
		return
	}

	customerIDStr, ok := customerIDVal.(string)
	if !ok || customerIDStr == "" {
		c.String(http.StatusUnauthorized, "Invalid customer ID")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid customer ID format")
		return
	}

	type completedProjectRow struct {
		ID               uuid.UUID
		ProjectRequestID *uuid.UUID
		Name             string
		Description      string
		Status           string
		DueDate          time.Time
		FinalQAStatus    string
		UpdatedAt        time.Time
	}

	var projects []completedProjectRow
	err = prc.Repo.DB.
		Table("projects").
		Select("projects.id, projects.project_request_id, projects.name, projects.description, projects.status, projects.due_date, projects.final_qa_status, projects.updated_at").
		Joins("JOIN project_requests ON project_requests.id = projects.project_request_id").
		Where("project_requests.customer_id = ? AND projects.status = ?", customerID, "completed").
		Order("projects.updated_at DESC").
		Scan(&projects).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching completed projects")
		return
	}

	c.HTML(http.StatusOK, "customer/projects/done.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Done Projects",
		"ActivePage": "projects-done",
		"projects":   projects,
	}))
}

// GET /customer/projects/create
func (prc *ProjectRequestController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "customer/projects/create.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Create Project",
		"ActivePage": "project-requests",
	}))
}

// POST /customer/projects/store
func (prc *ProjectRequestController) Store(c *gin.Context) {
	// Get customer ID from context (same as before)
	customerIDVal, exists := c.Get("customer_id")

	if !exists {
		c.String(http.StatusUnauthorized, "Customer not logged in")
		return
	}

	customerIDStr, ok := customerIDVal.(string)
	if !ok || customerIDStr == "" {
		c.String(http.StatusUnauthorized, "Invalid customer ID")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid customer ID format")
		return
	}

	// Bind form data
	var form struct {
		Title       string `form:"title" binding:"required"`
		Description string `form:"description"`
		Deadline    string `form:"deadline" binding:"required"`
	}

	// Bind form data
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Parse the date string into time.Time
	deadline, err := time.Parse("2006-01-02", form.Deadline)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid deadline format")
		return
	}

	// Create ProjectRequest with time.Time field
	project := models.ProjectRequest{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Title:       form.Title,
		Description: form.Description,
		Deadline:    deadline, // assign time.Time here
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := prc.Repo.Create(&project); err != nil {
		c.String(http.StatusInternalServerError, "Error creating project")
		return
	}

	c.Redirect(http.StatusSeeOther, "/customer/projects")

}
