package controllers

import (
	"net/http"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DepartmentController struct {
	Repo *repositories.DepartmentRepository
}

func NewDepartmentController(repo *repositories.DepartmentRepository) *DepartmentController {
	return &DepartmentController{Repo: repo}
}

// GET /admin/departments
func (dc *DepartmentController) Index(c *gin.Context) {
	departments, err := dc.Repo.All()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/departments/index.html", gin.H{
			"Error": "Failed to fetch departments",
		})
		return
	}

	c.HTML(http.StatusOK, "admin/departments/index.html", gin.H{
		"Departments": departments,
	})
}

// GET /admin/departments/create
func (dc *DepartmentController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/departments/create.html", nil)
}

// POST /admin/departments/store
func (dc *DepartmentController) Store(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.HTML(http.StatusBadRequest, "admin/departments/create.html", gin.H{
			"Error": "Department name is required",
		})
		return
	}

	department := &models.Department{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dc.Repo.Create(department); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/departments/create.html", gin.H{
			"Error": "Failed to create department",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/departments")
}

// GET /admin/departments/delete/:id
func (dc *DepartmentController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/departments/index.html", gin.H{
			"Error": "Invalid department ID",
		})
		return
	}

	if err := dc.Repo.Delete(id); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/departments/index.html", gin.H{
			"Error": "Failed to delete department",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/departments")
}

// GET /admin/departments/edit/:id
func (dc *DepartmentController) Edit(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/departments/index.html", gin.H{
			"Error": "Invalid department ID",
		})
		return
	}

	var department models.Department
	if err := dc.Repo.DB.First(&department, "id = ?", id).Error; err != nil {
		c.HTML(http.StatusNotFound, "admin/departments/index.html", gin.H{
			"Error": "Department not found",
		})
		return
	}

	c.HTML(http.StatusOK, "admin/departments/edit.html", gin.H{
		"Department": department,
	})
}

// POST /admin/departments/update/:id
func (dc *DepartmentController) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/departments/edit.html", gin.H{
			"Error": "Invalid department ID",
		})
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.HTML(http.StatusBadRequest, "admin/departments/edit.html", gin.H{
			"Error": "Department name is required",
		})
		return
	}

	var department models.Department
	if err := dc.Repo.DB.First(&department, "id = ?", id).Error; err != nil {
		c.HTML(http.StatusNotFound, "admin/departments/edit.html", gin.H{
			"Error": "Department not found",
		})
		return
	}

	department.Name = name
	department.UpdatedAt = time.Now()

	if err := dc.Repo.DB.Save(&department).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin/departments/edit.html", gin.H{
			"Error": "Failed to update department",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/departments")
}
