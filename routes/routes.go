package routes

import (
	"net/http"
	"work-management-system/config"
	"work-management-system/controllers"
	"work-management-system/middleware"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// --------------------
	// Global middleware
	// --------------------
	r.Use(middleware.Logger())

	// --------------------
	// Repositories
	// --------------------
	authRepo := repositories.NewAuthenticationRepository(config.DB)
	adminRepo := repositories.NewAdminRepository(config.DB)
	employeeRepo := repositories.NewEmployeeRepository(config.DB)
	customerRepo := repositories.NewCustomerRepository(config.DB)
	qaRepo := repositories.NewQualityAssuranceRepository(config.DB)
	managerRepo := repositories.NewManagerRepository(config.DB)

	roleRepo := repositories.NewRoleRepository(config.DB)
	permRepo := repositories.NewPermissionRepository(config.DB)
	deptRepo := repositories.NewDepartmentRepository(config.DB)
	userRepo := repositories.NewUserRepository(config.DB)
	teamRepo := repositories.NewTeamRepository(config.DB)
	projectRequestRepo := repositories.NewProjectRequestRepository(config.DB)
	taskRepo := repositories.NewTaskRepository(config.DB)
	projectRepo := repositories.NewProjectRepository(config.DB)
	eventRepo := repositories.NewEventRepository(config.DB)
	fileRepo := repositories.NewFileRepository(config.DB)
	taskReviewCommentRepo := repositories.NewTaskReviewCommentRepository(config.DB)
	notificationRepo := repositories.NewNotificationRepository(config.DB)
	activityLogRepo := repositories.NewActivityLogRepository(config.DB)
	reportRepo := repositories.NewReportRepository(config.DB)

	// --------------------
	// Controllers
	// --------------------
	authController := controllers.NewAuthenticationController(authRepo)
	adminController := controllers.NewAdminController(adminRepo, activityLogRepo)
	customerController := controllers.NewCustomerController(customerRepo, userRepo, notificationRepo)
	employeeController := controllers.NewEmployeeController(employeeRepo, userRepo, teamRepo, taskRepo, projectRepo, notificationRepo)
	qaController := controllers.NewQualityAssuranceController(qaRepo, userRepo, taskRepo, projectRepo, notificationRepo)
	managerController := controllers.NewManagerController(managerRepo, projectRepo, teamRepo, userRepo, notificationRepo)

	roleController := controllers.NewRoleController(roleRepo)
	permissionController := controllers.NewPermissionController(permRepo)
	departmentController := controllers.NewDepartmentController(deptRepo)
	userController := controllers.NewUserController(userRepo, roleRepo, deptRepo)
	teamController := controllers.NewTeamController(teamRepo, userRepo)
	projectController := controllers.NewProjectController(projectRepo, userRepo, teamRepo, eventRepo)
	projectRequestController := controllers.NewProjectRequestController(projectRequestRepo)
	taskController := controllers.NewTaskController(taskRepo, projectRepo, userRepo, eventRepo, notificationRepo)
	eventController := controllers.NewEventController(eventRepo, notificationRepo)
	fileController := controllers.NewFileController(fileRepo, taskRepo, projectRepo)
	supervisorController := controllers.NewSupervisorController(projectRepo, userRepo, notificationRepo)
	taskReviewController := controllers.NewTaskReviewController(taskRepo, projectRepo, taskReviewCommentRepo, notificationRepo)
	notificationController := controllers.NewNotificationController(notificationRepo)
	reportController := controllers.NewReportController(reportRepo)

	// --------------------
	// Public routes
	// --------------------
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "authentication/login.html", gin.H{
			"title": "Login",
		})
	})
	r.POST("/login", authController.Login)
	r.GET("/signup", authController.SignupPage)
	r.POST("/signup", authController.Signup)
	r.POST("/refresh", authController.RefreshToken)

	r.GET("/forgot-password", authController.ForgotPasswordPage)
	r.POST("/forgot-password", authController.ForgotPassword)

	r.GET("/reset-password", authController.ResetPasswordPage)
	r.POST("/reset-password", authController.ResetPassword)

	r.GET("/logout", middleware.AuthRequired(customerRepo), authController.Logout)

	// --------------------
	// Notifications API (all authenticated users)
	// --------------------
	api := r.Group("/api")
	api.Use(middleware.AuthRequired(customerRepo))
	{
		api.POST("/notifications/mark-read", notificationController.MarkRead)
		api.POST("/notifications/mark-all-read", notificationController.MarkAllRead)
	}

	// --------------------
	// Admin routes
	// --------------------
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("admin"), middleware.ActivityLogger(activityLogRepo))
	{
		//ROUTES
		admin.GET("/dashboard", adminController.Index)
		admin.GET("/profile", adminController.Profile)
		admin.GET("/notifications", adminController.Notifications)
		admin.GET("/activities", adminController.Activities)

		//EMPLOYEES
		employee := admin.Group("/employees")
		{
			employee.GET("/", userController.Index)
			employee.POST("/store", userController.Store)
			employee.POST("/update/:id", userController.Update)
			employee.POST("/delete/:id", userController.Delete)
		}

		//ROLES
		roles := admin.Group("/roles")
		{
			roles.GET("/", roleController.Index)
			roles.POST("/store", roleController.Store)
			roles.POST("/update/:id", roleController.Update)
			roles.POST("/delete/:id", roleController.Delete)

		}

		//DEPARTMENTS
		department := admin.Group("/departments")
		{
			department.GET("/", departmentController.Index)
			department.GET("/create", departmentController.Create)
			department.POST("/store", departmentController.Store)
			department.POST("/delete/:id", departmentController.Delete)
			department.GET("/edit/:id", departmentController.Edit)
			department.POST("/update/:id", departmentController.Update)
		}

		//TEAMS
		team := admin.Group("/teams")
		{
			team.GET("/", teamController.Index)
			team.GET("/chart", teamController.Chart)
			team.GET("/chart/api/orgchart", teamController.OrgChartData)
			team.GET("/create", teamController.Create)
			team.POST("/store", teamController.Store)
			team.GET("/edit/:id", teamController.Edit)
			team.POST("/update/:id", teamController.Update)
			team.POST("/delete/:id", teamController.Delete)
		}

		//PROJECTS
		project := admin.Group("/projects")
		{
			project.GET("/", projectController.Index)
			project.GET("/create", projectController.Create)
			project.POST("/store", projectController.Store)
			project.GET("/edit/:id", projectController.Edit)
			project.POST("/update/:id", projectController.Update)
			project.POST("/delete/:id", projectController.Delete)

			// Project requests
			project.GET("/requests", projectController.Requests)
			project.POST("/approve/:id", projectController.ApproveRequest)
			project.POST("/reject/:id", projectController.RejectRequest)

			// NEW: Approval workflow routes
			project.GET("/:id/approval", projectController.ShowApproval)
			project.POST("/:id/approve", projectController.ProcessApproval)
			project.POST("/:id/assign-approver", projectController.AssignApprover)
		}

		//TASKS
		task := admin.Group("/tasks")
		{
			task.GET("/", taskController.Index)
			task.GET("/create", taskController.Create)
			task.POST("/store", taskController.Store)
			task.GET("/edit/:id", taskController.Edit)
			task.POST("/update/:id", taskController.Update)
			task.POST("/delete/:id", taskController.Delete)
			task.GET("/board", taskController.Board)
			task.POST("/start/:id", taskController.StartTask)
			task.POST("/pause/:id", taskController.PauseTask)
			task.POST("/resume/:id", taskController.ResumeTask)
			task.POST("/upload/:taskId", fileController.Upload)
			task.GET("/files/:taskId", fileController.ListByTask)
			task.GET("/review-comments/:taskId", taskReviewController.ListComments)
			task.POST("/review/:taskId", taskReviewController.SubmitReview)
		}

		//CALENDAR EVENTS
		event := admin.Group("/events")
		{
			event.GET("/", eventController.Calendar)
			event.GET("/json", eventController.GetEventsJSON)
			event.POST("/store", eventController.Store)
			event.POST("/update/:id", eventController.Update)
			event.POST("/delete/:id", eventController.Delete)
		}

		permission := admin.Group("/permissions")
		{
			permission.GET("/", permissionController.Index)
			permission.POST("/store", permissionController.Store)
			permission.POST("/update/:id", permissionController.Update)
			permission.POST("/delete/:id", permissionController.Delete)
		}

		report := admin.Group("/reports")
		{
			report.GET("/employee", reportController.Employee)
			report.GET("/projects", reportController.Projects)
			report.GET("/project", reportController.Projects)
			report.GET("/task", reportController.Task)
		}
	}

	// --------------------
	// Manager routes
	// --------------------
	manager := r.Group("/manager")
	manager.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("manager"), middleware.ActivityLogger(activityLogRepo))
	{
		manager.GET("/dashboard", managerController.Index)
		manager.GET("/profile", managerController.Profile)

		task := manager.Group("/tasks")
		{
			task.GET("/board", taskController.Board)
			task.GET("/files/:taskId", fileController.ListByTask)
			task.GET("/review-comments/:taskId", taskReviewController.ListComments)
			task.POST("/review/:taskId", taskReviewController.SubmitReview)
		}

		project := manager.Group("/projects")
		{
			project.GET("/reviews", managerController.ProjectReviews)
			project.POST("/reviews/:id", managerController.ProcessProjectReview)
			project.POST("/:id/assign-approver", managerController.AssignApprover)
			project.GET("/requests", managerController.ProjectRequests)
			project.POST("/requests/:id/approve", managerController.ApproveProjectRequest)
			project.POST("/requests/:id/reject", managerController.RejectProjectRequest)
		}

		team := manager.Group("/teams")
		{
			team.GET("/", managerController.Teams)
			team.POST("/store", managerController.StoreTeam)
			team.POST("/update/:id", managerController.UpdateTeam)
			team.POST("/delete/:id", managerController.DeleteTeam)
		}

		event := manager.Group("/events")
		{
			event.GET("/", eventController.Calendar)
			event.GET("/json", eventController.GetEventsJSON)
			event.POST("/store", eventController.Store)
			event.POST("/update/:id", eventController.Update)
			event.POST("/delete/:id", eventController.Delete)
		}

		manager.GET("/notifications", managerController.Notifications)
	}

	// --------------------
	// Customer routes
	// --------------------
	customer := r.Group("/customer")
	customer.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("customer"), middleware.ActivityLogger(activityLogRepo))
	{
		customer.GET("/dashboard", customerController.Index)
		customer.GET("/profile", customerController.Profile)
		customer.GET("/notifications", customerController.Notifications)

		//PROJECT REQUESTS
		projects := customer.Group("/projects")
		{
			projects.GET("/", projectRequestController.Index)
			projects.GET("/done", projectRequestController.DoneProjects)
			projects.GET("/create", projectRequestController.Create)
			projects.POST("/store", projectRequestController.Store)
			projects.GET("/:requestId/files", fileController.ListCustomerRequestFiles)
		}
	}

	// --------------------
	// Customer routes
	// --------------------
	qa := r.Group("/qa")
	qa.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("qualityassurance"), middleware.ActivityLogger(activityLogRepo))
	{
		qa.GET("/dashboard", qaController.Index)
		qa.GET("/profile", qaController.Profile)
		qa.GET("/notifications", qaController.Notifications)

		//PROJECT REQUESTS
		projects := qa.Group("/projects")
		{
			projects.GET("/", projectRequestController.Index)
			projects.GET("/create", projectRequestController.Create)
			projects.POST("/store", projectRequestController.Store)
			projects.GET("/reviews", qaController.ProjectReviews)
			projects.POST("/reviews/:id", qaController.ProcessProjectReview)
			projects.POST("/send/:id", qaController.SendApprovedProject)
		}

		task := qa.Group("/tasks")
		{
			task.GET("/", taskController.Board)
			task.GET("/board", taskController.Board)
			task.GET("/files/:taskId", fileController.ListByTask)
			task.GET("/review-comments/:taskId", taskReviewController.ListComments)
			task.POST("/review/:taskId", taskReviewController.SubmitReview)
		}

		event := qa.Group("/events")
		{
			event.GET("/", eventController.Calendar)
			event.GET("/json", eventController.GetEventsJSON)
			event.POST("/store", eventController.Store)
			event.POST("/update/:id", eventController.Update)
			event.POST("/delete/:id", eventController.Delete)
		}
	}

	// --------------------
	// Employee routes
	// --------------------
	employee := r.Group("/employee")
	employee.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("employee"), middleware.ActivityLogger(activityLogRepo))
	{
		employee.GET("/dashboard", employeeController.Index)
		employee.GET("/profile", employeeController.Profile)
		employee.GET("/notifications", employeeController.Notifications)

		//CALENDAR EVENTS
		event := employee.Group("/events")
		{
			event.GET("/", eventController.Calendar)
			event.GET("/json", eventController.GetEventsJSON)
			event.POST("/store", eventController.Store)
			event.POST("/update/:id", eventController.Update)
			event.POST("/delete/:id", eventController.Delete)
		}

		//TASKS
		task := employee.Group("/tasks")
		{
			task.GET("/", taskController.Index)
			task.GET("/create", taskController.Create)
			task.POST("/store", taskController.Store)
			task.GET("/edit/:id", taskController.Edit)
			task.POST("/update/:id", taskController.Update)
			task.GET("/board", taskController.Board)
			task.POST("/start/:id", taskController.StartTask)
			task.POST("/pause/:id", taskController.PauseTask)
			task.POST("/resume/:id", taskController.ResumeTask)
			task.POST("/upload/:taskId", fileController.Upload)
			task.GET("/files/:taskId", fileController.ListByTask)
			task.GET("/review-comments/:taskId", taskReviewController.ListComments)
		}

		//TEAMS
		team := employee.Group("/teams")
		{
			team.GET("/", teamController.Index)
			team.GET("/chart", teamController.Chart)
			team.GET("/chart/api/orgchart", teamController.OrgChartData)
		}

	}

	// --------------------
	// Supervisor routes
	// --------------------
	supervisor := r.Group("/supervisor")
	supervisor.Use(middleware.AuthRequired(customerRepo), middleware.RequireRole("supervisor"), middleware.ActivityLogger(activityLogRepo))
	{
		supervisor.GET("/dashboard", supervisorController.Index)
		supervisor.GET("/profile", supervisorController.Profile)
		supervisor.GET("/notifications", supervisorController.Notifications)

		task := supervisor.Group("/tasks")
		{
			task.GET("/", taskController.Index)
			task.POST("/store", taskController.Store)
			task.GET("/board", taskController.Board)
			task.POST("/start/:id", taskController.StartTask)
			task.POST("/pause/:id", taskController.PauseTask)
			task.POST("/resume/:id", taskController.ResumeTask)
			task.GET("/files/:taskId", fileController.ListByTask)
			task.GET("/review-comments/:taskId", taskReviewController.ListComments)
		}

		project := supervisor.Group("/projects")
		{
			project.GET("/", supervisorController.Projects)
			project.POST("/:id/assign-employee", supervisorController.AssignProjectEmployee)
			project.POST("/submit-for-qa/:id", supervisorController.SubmitProjectForQA)
		}

		event := supervisor.Group("/events")
		{
			event.GET("/", eventController.Calendar)
			event.GET("/json", eventController.GetEventsJSON)
			event.POST("/store", eventController.Store)
			event.POST("/update/:id", eventController.Update)
			event.POST("/delete/:id", eventController.Delete)
		}

		team := supervisor.Group("/teams")
		{
			team.GET("/", teamController.Index)
			team.GET("/chart", teamController.Chart)
			team.GET("/chart/api/orgchart", teamController.OrgChartData)
		}
	}

	// --------------------
	// 404 fallback
	// --------------------
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "errors/error.html", gin.H{})
	})

}
