package controllers

import (
	"net/http"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	Repo *repositories.ReportRepository
}

func NewReportController(repo *repositories.ReportRepository) *ReportController {
	return &ReportController{Repo: repo}
}

func (rc *ReportController) Employee(c *gin.Context) {
	report, err := rc.Repo.EmployeeReport()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load employee report")
		return
	}

	c.HTML(http.StatusOK, "admin/reports/employee.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Employee Reports",
		"ActivePage": "employee-reports",
		"Report":     report,
	}))
}

func (rc *ReportController) Projects(c *gin.Context) {
	report, err := rc.Repo.ProjectReport()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load project report")
		return
	}

	c.HTML(http.StatusOK, "admin/reports/projects.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Project Reports",
		"ActivePage": "project-reports",
		"Report":     report,
	}))
}

func (rc *ReportController) Task(c *gin.Context) {
	report, err := rc.Repo.TaskReport()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load task report")
		return
	}

	c.HTML(http.StatusOK, "admin/reports/task.html", utils.TemplateContext(c, gin.H{
		"PageTitle":  "Task Reports",
		"ActivePage": "task-reports",
		"Report":     report,
	}))
}
