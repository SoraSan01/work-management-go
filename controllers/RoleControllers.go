package controllers

import (
	"net/http"

	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleController struct {
	Repo *repositories.RoleRepository
}

func NewRoleController(repo *repositories.RoleRepository) *RoleController {
	return &RoleController{Repo: repo}
}

// GET /admin/roles
func (rc *RoleController) Index(c *gin.Context) {
	roles, err := rc.Repo.All()
	perms, _ := rc.Repo.AllPermissions()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/roles/list.html", utils.TemplateContext(c, gin.H{
			"Error": "Failed to load roles",
		}))
		return
	}

	c.HTML(http.StatusOK, "admin/roles/list.html", utils.TemplateContext(c, gin.H{
		"Roles":       roles,
		"Permissions": perms,
	}))
}

// POST /admin/roles/store
func (rc *RoleController) Store(c *gin.Context) {
	name := c.PostForm("name")
	permissionIDs := c.PostFormArray("permissions")

	if exists, _ := rc.Repo.ExistsByName(name); exists {
		perms, _ := rc.Repo.AllPermissions()
		c.HTML(http.StatusBadRequest, "admin/roles/create.html", utils.TemplateContext(c, gin.H{
			"Error":       "Role already exists",
			"Permissions": perms,
		}))
		return
	}

	role := &models.Role{Name: name}
	if err := rc.Repo.Create(role, permissionIDs); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/roles/create.html", utils.TemplateContext(c, gin.H{
			"Error": "Failed to create role",
		}))
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/roles")
}

// POST /admin/roles/update/:id
func (rc *RoleController) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	name := c.PostForm("name")
	permissionIDs := c.PostFormArray("permissions")

	role := &models.Role{ID: id, Name: name}
	if err := rc.Repo.Update(role, permissionIDs); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/roles/edit.html", utils.TemplateContext(c, gin.H{
			"Error": "Failed to update role",
		}))
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/roles")
}

// GET /admin/roles/delete/:id
func (rc *RoleController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/roles/list.html", utils.TemplateContext(c, gin.H{
			"Error": "Invalid role ID",
		}))
		return
	}

	if err := rc.Repo.Delete(id); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/roles/list.html", utils.TemplateContext(c, gin.H{
			"Error": "Failed to delete role",
		}))
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/roles")
}
