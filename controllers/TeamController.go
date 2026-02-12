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

type TeamController struct {
	Repo     *repositories.TeamRepository
	UserRepo *repositories.UserRepository
}

func NewTeamController(repo *repositories.TeamRepository, userRepo *repositories.UserRepository) *TeamController {
	return &TeamController{
		Repo:     repo,
		UserRepo: userRepo,
	}
}

type userOption struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	TeamID    *uuid.UUID
}

type OrgNode struct {
	ID           string `json:"id"`
	ParentID     string `json:"parentId"`
	Name         string `json:"name"`
	PositionName string `json:"positionName"`
	ImageURL     string `json:"imageUrl"`
}

// List all teams
func (tc *TeamController) Index(c *gin.Context) {
	role := c.GetString("role")
	templatePath := "employee/teams/index.html"
	activePage := "teams"
	var teams []models.Team
	var err error

	supervisors, _ := tc.Repo.GetSupervisors()
	employees, _ := tc.Repo.GetEmployees()
	users, _ := tc.Repo.GetAllUsers()

	teamByUser := make(map[uuid.UUID]uuid.UUID)

	if role == "admin" {
		teams, err = tc.Repo.GetAll()
	} else {
		uidStr := c.GetString("user_id")
		uid, _ := uuid.Parse(uidStr)

		teams, err = tc.Repo.GetByUserID(uid)
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/teams/index.html", gin.H{
			"Error": "Failed to load teams",
		})
		return
	}

	for _, team := range teams {
		for _, member := range team.Members {
			teamByUser[member.ID] = team.ID
		}
	}

	buildOptions := func(list []models.User) []userOption {
		opts := make([]userOption, 0, len(list))
		for _, u := range list {
			var teamID *uuid.UUID
			if id, ok := teamByUser[u.ID]; ok {
				tmp := id
				teamID = &tmp
			}
			opts = append(opts, userOption{
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				TeamID:    teamID,
			})
		}
		return opts
	}

	if role == "admin" {
		templatePath = "admin/teams/index.html"
	} else if role == "supervisor" {
		templatePath = "supervisor/teams/index.html"
	}

	c.HTML(http.StatusOK, templatePath, utils.TemplateContext(c, gin.H{
		"ActivePage":          activePage,
		"Teams":               teams,
		"Supervisors":         supervisors,
		"Employees":           employees,
		"Users":               users,
		"SupervisorsWithTeam": buildOptions(supervisors),
		"UsersWithTeam":       buildOptions(users),
	}))
}

// Org chart view
func (tc *TeamController) Chart(c *gin.Context) {
	teams, err := tc.Repo.GetAll()
	managers, _ := tc.UserRepo.GetUsersByRole("Manager")
	role := c.GetString("role")
	templatePath := "employee/teams/chart.html"

	if err != nil {
		if role == "admin" {
			templatePath = "admin/teams/chart.html"
		}
		c.HTML(http.StatusInternalServerError, templatePath, gin.H{
			"Error": "Failed to load teams",
		})
		return
	}

	if role == "admin" {
		templatePath = "admin/teams/chart.html"
	} else if role == "supervisor" {
		templatePath = "supervisor/teams/chart.html"
	}

	c.HTML(http.StatusOK, templatePath, utils.TemplateContext(c, gin.H{
		"ActivePage": "team-chart",
		"Teams":    teams,
		"Managers": managers,
	}))
}

// Show create team form
func (tc *TeamController) Create(c *gin.Context) {
	users, _ := tc.Repo.GetAllUsers()
	supervisors, _ := tc.Repo.GetSupervisors()
	employees, _ := tc.Repo.GetEmployees()

	c.HTML(http.StatusOK, "admin/teams/create.html", gin.H{
		"Users":       users,
		"Supervisors": supervisors,
		"Employees":   employees,
	})
}

// Store new team
func (tc *TeamController) Store(c *gin.Context) {
	name := c.PostForm("name")
	supervisorID := c.PostForm("supervisor_id")
	memberIDs := c.PostFormArray("member_ids")
	if len(memberIDs) == 0 {
		memberIDs = c.PostFormArray("member_ids[]")
	}

	// Validate required fields
	if name == "" || supervisorID == "" {
		supervisors, _ := tc.Repo.GetSupervisors()
		employees, _ := tc.Repo.GetEmployees()
		c.HTML(http.StatusBadRequest, "admin/teams/create.html", gin.H{
			"Error":       "Team name and supervisor are required",
			"Supervisors": supervisors,
			"Employees":   employees,
		})
		return
	}

	// Parse supervisor ID
	supervisorUUID, err := uuid.Parse(supervisorID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/teams/create.html", gin.H{
			"Error": "Invalid supervisor ID",
		})
		return
	}

	// Create team struct
	team := &models.Team{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add supervisor as a member automatically
	team.Members = append(team.Members, models.User{ID: supervisorUUID})
	team.SupervisorID = supervisorUUID

	// Add selected members
	for _, idStr := range memberIDs {
		id, err := uuid.Parse(idStr)
		if err == nil && id != supervisorUUID { // prevent duplicate
			team.Members = append(team.Members, models.User{ID: id})
		}
	}

	// Create team
	if err := tc.Repo.Create(team); err != nil {
		supervisors, _ := tc.Repo.GetSupervisors()
		employees, _ := tc.Repo.GetEmployees()
		c.HTML(http.StatusInternalServerError, "admin/teams/create.html", gin.H{
			"Error":       "Failed to create team",
			"Supervisors": supervisors,
			"Employees":   employees,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/teams")
}

// Show edit form
func (tc *TeamController) Edit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/teams")
		return
	}

	team, err := tc.Repo.GetByID(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/teams")
		return
	}

	users, _ := tc.Repo.GetAllUsers()
	c.HTML(http.StatusOK, "admin/teams/edit.html", gin.H{
		"Team":  team,
		"Users": users,
	})
}

// Update team
func (tc *TeamController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/teams")
		return
	}

	team, err := tc.Repo.GetByID(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/teams")
		return
	}

	name := c.PostForm("name")
	supervisorID := c.PostForm("supervisor_id")
	memberIDs := c.PostFormArray("member_ids")
	if len(memberIDs) == 0 {
		memberIDs = c.PostFormArray("member_ids[]")
	}

	team.Name = name
	team.UpdatedAt = time.Now()
	team.Members = nil

	if supervisorID != "" {
		supervisorUUID, err := uuid.Parse(supervisorID)
		if err == nil {
			team.SupervisorID = supervisorUUID
			team.Members = append(team.Members, models.User{ID: supervisorUUID})
		}
	}

	for _, idStr := range memberIDs {
		mid, err := uuid.Parse(idStr)
		if err == nil {
			if team.SupervisorID != uuid.Nil && mid == team.SupervisorID {
				continue
			}
			team.Members = append(team.Members, models.User{ID: mid})
		}
	}

	if err := tc.Repo.Update(team); err != nil {
		users, _ := tc.Repo.GetAllUsers() // capture both values
		c.HTML(http.StatusInternalServerError, "admin/teams/edit.html", gin.H{
			"Error": "Failed to update team",
			"Team":  team,
			"Users": users, // now it's a single value
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/teams")
}

// Delete a team
func (tc *TeamController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err == nil {
		tc.Repo.Delete(id)
	}
	c.Redirect(http.StatusSeeOther, "/admin/teams")
}

func (tc *TeamController) OrgChartData(c *gin.Context) {
	role := c.GetString("role")
	var teams []models.Team

	if role == "supervisor" {
		uid, err := uuid.Parse(c.GetString("user_id"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid supervisor"})
			return
		}
		teams, _ = tc.Repo.GetByUserID(uid)
	} else {
		teams, _ = tc.Repo.GetAll()
	}

	var nodes []OrgNode

	rootID := "root"

	// single root
	nodes = append(nodes, OrgNode{
		ID:           rootID,
		ParentID:     "",
		Name:         "Organization",
		PositionName: "",
	})

	for _, team := range teams {

		teamID := team.ID.String()

		// Team under root
		nodes = append(nodes, OrgNode{
			ID:           teamID,
			ParentID:     rootID,
			Name:         team.Name,
			PositionName: "Team",
		})

		// Supervisor
		if team.SupervisorID != uuid.Nil {
			nodes = append(nodes, OrgNode{
				ID:           team.Supervisor.ID.String(),
				ParentID:     teamID,
				Name:         team.Supervisor.FirstName + " " + team.Supervisor.LastName,
				PositionName: "Supervisor",
			})
		}

		// Members
		for _, m := range team.Members {
			if m.ID == team.SupervisorID {
				continue
			}

			nodes = append(nodes, OrgNode{
				ID:           m.ID.String(),
				ParentID:     teamID,
				Name:         m.FirstName + " " + m.LastName,
				PositionName: "Member",
			})
		}
	}

	c.JSON(http.StatusOK, nodes)
}
