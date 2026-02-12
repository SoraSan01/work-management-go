package controllers

import (
	"net/http"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PermissionController struct {
	Repo *repositories.PermissionRepository
}

func NewPermissionController(repo *repositories.PermissionRepository) *PermissionController {
	return &PermissionController{Repo: repo}
}

func (pc *PermissionController) Index(c *gin.Context) {
	permissions, err := pc.Repo.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error/error.html", gin.H{"Message": "Failed to retrieve permissions."})
		return
	}
	c.HTML(http.StatusOK, "admin/permissions/list.html", gin.H{"Permissions": permissions})
}

func (pc *PermissionController) Store(c *gin.Context) {
	var permission models.Permission
	if err := c.ShouldBind(&permission); err != nil {
		c.HTML(http.StatusBadRequest, "error/error.html", gin.H{"Message": "Invalid input."})
		return
	}
	if err := pc.Repo.Create(&permission); err != nil {
		c.HTML(http.StatusInternalServerError, "error/error.html", gin.H{"Message": "Failed to create permission."})
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/permissions")
}

func (pc *PermissionController) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	name := c.PostForm("name")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error/error.html", gin.H{"Message": "Invalid permission ID."})
		return
	}
	permission, err := pc.Repo.GetByID(id)
	permission.Name = name
	if err != nil {
		c.HTML(http.StatusNotFound, "error/error.html", gin.H{"Message": "Permission not found."})
		return
	}

	if err := pc.Repo.Update(permission); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/permissions/edit.html", utils.TemplateContext(c, gin.H{
			"Error": "Failed to update permission",
		}))
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/permissions")
}

func (pc *PermissionController) Delete(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if err := pc.Repo.Delete(uid); err != nil {
		c.HTML(http.StatusInternalServerError, "error/error.html", gin.H{
			"Message": "Failed to delete permission.",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/permissions")
}
