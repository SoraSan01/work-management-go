package controllers

import (
	"net/http"
	"work-management-system/models"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	Repo     *repositories.UserRepository
	RoleRepo *repositories.RoleRepository
	DepRepo  *repositories.DepartmentRepository
}

func NewUserController(repo *repositories.UserRepository, rr *repositories.RoleRepository, dr *repositories.DepartmentRepository) *UserController {
	return &UserController{
		Repo:     repo,
		RoleRepo: rr,
		DepRepo:  dr,
	}
}

// GET /employees
func (uc *UserController) Index(c *gin.Context) {
	users, err := uc.Repo.All()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/employees/index.html", gin.H{
			"Error": "Failed to fetch employees",
		})
		return
	}

	roles, _ := uc.RoleRepo.All()
	departments, _ := uc.DepRepo.All()

	c.HTML(http.StatusOK, "admin/employees/index.html", gin.H{
		"Employees":   users,
		"Roles":       roles,
		"Departments": departments,
	})
}

// POST /employees/store
func (uc *UserController) Store(c *gin.Context) {
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")
	roleIDStr := c.PostForm("role_id")
	depIDStr := c.PostForm("department_id")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if password != confirmPassword {
		roles, _ := uc.RoleRepo.All()
		departments, _ := uc.DepRepo.All()
		c.HTML(http.StatusBadRequest, "admin/employees/index.html", gin.H{
			"Error":       "Passwords do not match",
			"Roles":       roles,
			"Departments": departments,
		})
		return
	}

	// Check if email already exists
	exists, _ := uc.Repo.ExistsByEmail(email)
	if exists {
		roles, _ := uc.RoleRepo.All()
		departments, _ := uc.DepRepo.All()
		c.HTML(http.StatusBadRequest, "admin/employees/index.html", gin.H{
			"Error":       "Email already exists",
			"Roles":       roles,
			"Departments": departments,
		})
		return
	}

	// Convert RoleID and DepartmentID to uuid.UUID
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Role ID")
		return
	}

	var depID *uuid.UUID

	if depIDStr != "" {
		parsed, err := uuid.Parse(depIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Department ID")
			return
		}
		depID = &parsed
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		RoleID:       roleID,
		DepartmentID: depID,
	}

	if err := uc.Repo.Create(user, password); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/employees/index.html", gin.H{
			"Error": "Failed to create employee",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/employees")
}

// GET /employees/edit/:id
func (uc *UserController) Edit(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.Repo.FindByID(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/employees")
		return
	}

	roles, _ := uc.RoleRepo.All()
	departments, _ := uc.DepRepo.All()

	c.HTML(http.StatusOK, "admin/employees/edit.html", gin.H{
		"Employee":    user,
		"Roles":       roles,
		"Departments": departments,
	})
}

// POST /employees/edit/:id
func (uc *UserController) Update(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.Repo.FindByID(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/employees")
		return
	}

	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")
	roleIDStr := c.PostForm("role_id")
	depIDStr := c.PostForm("department_id")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Password validation
	if password != "" && password != confirmPassword {
		roles, _ := uc.RoleRepo.All()
		departments, _ := uc.DepRepo.All()
		c.HTML(http.StatusBadRequest, "admin/employees/edit.html", gin.H{
			"Error":       "Passwords do not match",
			"Employee":    user,
			"Roles":       roles,
			"Departments": departments,
		})
		return
	}

	// Convert RoleID and DepartmentID to uuid.UUID
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Role ID")
		return
	}

	var depID *uuid.UUID

	if depIDStr != "" {
		parsed, err := uuid.Parse(depIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Department ID")
			return
		}
		depID = &parsed
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.RoleID = roleID
	user.DepartmentID = depID

	if err := uc.Repo.Update(user, password); err != nil {
		roles, _ := uc.RoleRepo.All()
		departments, _ := uc.DepRepo.All()
		c.HTML(http.StatusInternalServerError, "admin/employees/edit.html", gin.H{
			"Error":       "Failed to update employee",
			"Employee":    user,
			"Roles":       roles,
			"Departments": departments,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/employees")
}

// GET /employees/delete/:id
func (uc *UserController) Delete(c *gin.Context) {
	id := c.Param("id")
	_ = uc.Repo.Delete(id)
	c.Redirect(http.StatusSeeOther, "/admin/employees")
}
