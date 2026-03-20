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

func (uc *UserController) renderIndex(c *gin.Context, status int, data gin.H) {
	users, _ := uc.Repo.All()
	roles, _ := uc.RoleRepo.All()
	departments, _ := uc.DepRepo.All()
	managers, _ := uc.Repo.GetUsersByRole("Manager")

	if data == nil {
		data = gin.H{}
	}

	data["Employees"] = users
	data["Roles"] = roles
	data["Departments"] = departments
	data["Managers"] = managers

	c.HTML(status, "admin/employees/index.html", data)
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
		uc.renderIndex(c, http.StatusInternalServerError, gin.H{
			"Error": "Failed to fetch employees",
		})
		return
	}

	roles, _ := uc.RoleRepo.All()
	departments, _ := uc.DepRepo.All()
	managers, _ := uc.Repo.GetUsersByRole("Manager")

	c.HTML(http.StatusOK, "admin/employees/index.html", gin.H{
		"Employees":   users,
		"Roles":       roles,
		"Departments": departments,
		"Managers":    managers,
	})
}

// POST /employees/store
func (uc *UserController) Store(c *gin.Context) {
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")
	nationality := c.PostForm("nationality")
	roleIDStr := c.PostForm("role_id")
	depIDStr := c.PostForm("department_id")
	managerIDStr := c.PostForm("manager_id")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if password != confirmPassword {
		uc.renderIndex(c, http.StatusBadRequest, gin.H{
			"Error": "Passwords do not match",
		})
		return
	}

	// Check if email already exists
	exists, _ := uc.Repo.ExistsByEmail(email)
	if exists {
		uc.renderIndex(c, http.StatusBadRequest, gin.H{
			"Error": "Email already exists",
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
	var managerID *uuid.UUID

	if depIDStr != "" {
		parsed, err := uuid.Parse(depIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Department ID")
			return
		}
		depID = &parsed
	}

	if managerIDStr != "" {
		parsed, err := uuid.Parse(managerIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Manager ID")
			return
		}
		managerID = &parsed
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Nationality:  nationality,
		RoleID:       roleID,
		DepartmentID: depID,
		ManagerID:    managerID,
	}

	if err := uc.Repo.Create(user, password); err != nil {
		uc.renderIndex(c, http.StatusInternalServerError, gin.H{
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
	nationality := c.PostForm("nationality")
	roleIDStr := c.PostForm("role_id")
	depIDStr := c.PostForm("department_id")
	managerIDStr := c.PostForm("manager_id")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Password validation
	if password != "" && password != confirmPassword {
		uc.renderIndex(c, http.StatusBadRequest, gin.H{
			"Error":    "Passwords do not match",
			"Employee": user,
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
	var managerID *uuid.UUID

	if depIDStr != "" {
		parsed, err := uuid.Parse(depIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Department ID")
			return
		}
		depID = &parsed
	}

	if managerIDStr != "" {
		parsed, err := uuid.Parse(managerIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid Manager ID")
			return
		}
		managerID = &parsed
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Nationality = nationality
	user.RoleID = roleID
	user.DepartmentID = depID
	user.ManagerID = managerID

	if err := uc.Repo.Update(user, password); err != nil {
		uc.renderIndex(c, http.StatusInternalServerError, gin.H{
			"Error":    "Failed to update employee",
			"Employee": user,
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
