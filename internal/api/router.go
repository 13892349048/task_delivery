package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/api/handlers"
	"taskmanage/internal/api/middleware"
	"taskmanage/internal/container"
)

// NewRouter 创建新的路由器
func NewRouter(container *container.ApplicationContainer, logger *logrus.Logger) *gin.Engine {
	// 设置Gin模式 - 开发环境使用DebugMode以便看到更多日志
	if container.GetConfig().App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 设置全局中间件
	setupMiddleware(engine, logger)

	// 设置路由
	setupRoutes(engine, container, logger)

	// 添加启动完成日志
	logger.Info("路由器创建完成，所有路由已注册")

	return engine
}

// setupMiddleware 设置全局中间件
func setupMiddleware(engine *gin.Engine, logger *logrus.Logger) {
	// 恢复中间件
	engine.Use(gin.Recovery())

	// 请求ID中间件（必须在日志中间件之前）
	engine.Use(middleware.RequestID())

	// 日志中间件
	engine.Use(middleware.Logger(logger))

	// CORS中间件
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 超时中间件
	engine.Use(middleware.Timeout(30 * time.Second))
	
	logger.Info("中间件设置完成")
}

// setupRoutes 设置路由
func setupRoutes(engine *gin.Engine, container *container.ApplicationContainer, logger *logrus.Logger) {
	// 健康检查路由
	setupHealthRoutes(engine, container, logger)

	// API路由组
	setupAPIRoutes(engine, container, logger)

}

// setupHealthRoutes 设置健康检查路由
func setupHealthRoutes(engine *gin.Engine, container *container.ApplicationContainer, logger *logrus.Logger) {
	healthHandler := handlers.NewHealthHandler(container, logger)

	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Task Management System",
			"status":  "running",
			"version": "1.0.0",
		})
	})

	engine.GET("/health", healthHandler.HealthCheck)
	engine.GET("/health/ready", healthHandler.ReadinessCheck)
	engine.GET("/health/live", healthHandler.LivenessCheck)
}

// setupAPIRoutes 设置API路由组
func setupAPIRoutes(engine *gin.Engine, container *container.ApplicationContainer, logger *logrus.Logger) {
	// 创建处理器
	authHandler := handlers.NewAuthHandler(container, logger)
	userHandler := handlers.NewUserHandler(container, logger)

	logger.Info("开始创建TaskHandler...")
	taskHandler := handlers.NewTaskHandler(container, logger)
	logger.Info("TaskHandler创建成功")
	assignmentHandler := handlers.NewAssignmentHandler(container, logger)
	employeeHandler := handlers.NewEmployeeHandler(container, logger)
	skillHandler := handlers.NewSkillHandler(container)
	notificationHandler := handlers.NewNotificationHandler(container, logger)
	workflowHandler := handlers.NewWorkflowHandler(container.GetServiceManager().WorkflowService(), logger)
	
	// 组织架构管理处理器
	departmentHandler := handlers.NewDepartmentHandler(container.GetServiceManager().DepartmentService(), logger)
	positionHandler := handlers.NewPositionHandler(container.GetServiceManager().PositionService(), logger)
	projectHandler := handlers.NewProjectHandler(container.GetServiceManager().ProjectService(), logger)
	
	// 权限分配处理器
	permissionAssignmentHandler := handlers.NewPermissionAssignmentHandler(container.GetServiceManager().PermissionAssignmentService(), logger)

	// API v1 路由组
	v1 := engine.Group("/api/v1")

	// 认证路由（无需认证）
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	// 需要认证的路由
	authenticated := v1.Group("/")
	authenticated.Use(middleware.Auth(container))

	// 用户管理路由
	users := authenticated.Group("/users")
	{
		users.GET("", middleware.RequirePermission(container, "user", "read"), userHandler.ListUsers)
		users.GET("/:id", middleware.RequirePermission(container, "user", "read"), userHandler.GetUser)
		users.PUT("/:id", middleware.RequirePermission(container, "user", "update"), userHandler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission(container, "user", "delete"), userHandler.DeleteUser)
		users.POST("/:id/roles", middleware.RequirePermission(container, "user", "assign_role"), userHandler.AssignRoles)
		users.DELETE("/:id/roles", middleware.RequirePermission(container, "user", "assign_role"), userHandler.RemoveRoles)
	}

	// 任务管理路由
	tasks := authenticated.Group("/tasks")
	{
		tasks.GET("", middleware.RequirePermission(container, "task", "read"), taskHandler.ListTasks)
		tasks.POST("", middleware.RequirePermission(container, "task", "create"), taskHandler.CreateTask)
		tasks.GET("/:id", middleware.RequirePermission(container, "task", "read"), taskHandler.GetTask)
		tasks.PUT("/:id", middleware.RequirePermission(container, "task", "update"), taskHandler.UpdateTask)
		tasks.DELETE("/:id", middleware.RequirePermission(container, "task", "delete"), taskHandler.DeleteTask)
		tasks.POST("/:id/assign", middleware.RequirePermission(container, "task", "assign"), taskHandler.AssignTask)
		tasks.POST("/:id/reassign", middleware.RequirePermission(container, "task", "assign"), taskHandler.ReassignTask)
		tasks.POST("/:id/start", middleware.RequirePermission(container, "task", "update"), taskHandler.StartTask)
		tasks.POST("/:id/complete", middleware.RequirePermission(container, "task", "update"), taskHandler.CompleteTask)
		tasks.POST("/:id/cancel", middleware.RequirePermission(container, "task", "update"), taskHandler.CancelTask)
		tasks.POST("/:id/auto-assign", middleware.RequirePermission(container, "task", "assign"), taskHandler.AutoAssignTask)
		tasks.GET("/:id/suggestions", middleware.RequirePermission(container, "task", "assign"), taskHandler.GetAssignmentSuggestions)
	}

	// 分配管理路由
	assignments := authenticated.Group("/assignments")
	{
		assignments.POST("/manual", middleware.RequirePermission(container, "task", "assign"), assignmentHandler.ManualAssign)
		assignments.POST("/suggestions", middleware.RequirePermission(container, "task", "assign"), assignmentHandler.GetAssignmentSuggestions)
		assignments.POST("/conflicts/:task_id", middleware.RequirePermission(container, "task", "assign"), assignmentHandler.CheckAssignmentConflicts)
		assignments.GET("/history/:task_id", middleware.RequirePermission(container, "task", "read"), assignmentHandler.GetAssignmentHistory)
		assignments.POST("/reassign/:task_id", middleware.RequirePermission(container, "task", "assign"), assignmentHandler.ReassignTask)
		assignments.POST("/cancel/:task_id", middleware.RequirePermission(container, "task", "assign"), assignmentHandler.CancelAssignment)
		assignments.GET("/strategies", middleware.RequirePermission(container, "task", "read"), assignmentHandler.GetAssignmentStrategies)
		assignments.GET("/stats", middleware.RequirePermission(container, "task", "read"), assignmentHandler.GetAssignmentStats)
	}

	// 员工管理路由
	employees := authenticated.Group("/employees")
	{
		employees.GET("", middleware.RequirePermission(container, "employee", "read"), employeeHandler.ListEmployees)
		employees.POST("", middleware.RequirePermission(container, "employee", "create"), employeeHandler.CreateEmployee)
		employees.GET("/:id", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetEmployee)
		employees.PUT("/:id", middleware.RequirePermission(container, "employee", "update"), employeeHandler.UpdateEmployee)
		employees.DELETE("/:id", middleware.RequirePermission(container, "employee", "delete"), employeeHandler.DeleteEmployee)
		employees.GET("/available", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetAvailableEmployees)
		employees.GET("/:id/workload", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetEmployeeWorkload)
		employees.POST("/:id/skills", middleware.RequirePermission(container, "employee", "update"), employeeHandler.AddSkill)
		employees.DELETE("/:id/skills", middleware.RequirePermission(container, "employee", "update"), employeeHandler.RemoveSkill)

		// 员工状态管理
		employees.PUT("/:id/status", middleware.RequirePermission(container, "employee", "update"), employeeHandler.UpdateEmployeeStatus)
		employees.GET("/status", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetEmployeesByStatus)

		// 工作负载统计
		employees.GET("/workload/stats", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetWorkloadStats)
		employees.GET("/workload/departments/:department", middleware.RequirePermission(container, "employee", "read"), employeeHandler.GetDepartmentWorkload)
	}

	// 通知路由
	notificationRoutes := v1.Group("/notifications")
	notificationRoutes.Use(middleware.Auth(container))
	{
		notificationRoutes.GET("", middleware.RequirePermission(container, "notification", "read"), notificationHandler.GetNotifications)
		notificationRoutes.GET("/count", middleware.RequirePermission(container, "notification", "read"), notificationHandler.GetUnreadCount)
		notificationRoutes.PUT("/:id/read", middleware.RequirePermission(container, "notification", "read"), notificationHandler.MarkAsRead)
		notificationRoutes.PUT("/read", middleware.RequirePermission(container, "notification", "read"), notificationHandler.MarkAllAsRead)
		notificationRoutes.POST("/:id/accept", middleware.RequirePermission(container, "task", "update"), notificationHandler.AcceptTask)
		notificationRoutes.POST("/:id/reject", middleware.RequirePermission(container, "task", "update"), notificationHandler.RejectTask)
		notificationRoutes.POST("/send", middleware.RequirePermission(container, "notification", "send"), notificationHandler.SendNotification)
		notificationRoutes.POST("/broadcast", middleware.RequirePermission(container, "notification", "send"), notificationHandler.BroadcastNotification)
	}

	// 工作流路由
	workflowRoutes := v1.Group("/workflows")
	workflowRoutes.Use(middleware.Auth(container))
	{
		// 工作流定义管理
		workflowRoutes.POST("/definitions", middleware.RequirePermission(container, "system", "admin"), workflowHandler.CreateWorkflowDefinition)
		workflowRoutes.GET("/definitions", middleware.RequirePermission(container, "task", "read"), workflowHandler.GetWorkflowDefinitions)
		workflowRoutes.GET("/definitions/:id", middleware.RequirePermission(container, "task", "read"), workflowHandler.GetWorkflowDefinition)
		workflowRoutes.PUT("/definitions/:id", middleware.RequirePermission(container, "system", "admin"), workflowHandler.UpdateWorkflowDefinition)
		workflowRoutes.DELETE("/definitions/:id", middleware.RequirePermission(container, "system", "admin"), workflowHandler.DeleteWorkflowDefinition)
		workflowRoutes.POST("/definitions/validate", middleware.RequirePermission(container, "system", "admin"), workflowHandler.ValidateWorkflowDefinition)
		
		// 任务分配审批流程
		workflowRoutes.POST("/task-assignment/start", middleware.RequirePermission(container, "task", "approve"), workflowHandler.StartTaskAssignmentApproval)
		
		// 审批处理
		workflowRoutes.POST("/approvals/process", middleware.RequirePermission(container, "task", "approve"), workflowHandler.ProcessApproval)
		workflowRoutes.GET("/approvals/pending", middleware.RequirePermission(container, "task", "approve"), workflowHandler.GetPendingApprovals)
		workflowRoutes.GET("/approvals/task-assignments", middleware.RequirePermission(container, "task", "approve"), workflowHandler.GetPendingTaskAssignmentApprovals)
		workflowRoutes.GET("/approvals/count", middleware.RequirePermission(container, "task", "approve"), workflowHandler.GetApprovalCount)
		
		// 流程实例管理
		workflowRoutes.GET("/instances/:instance_id", middleware.RequirePermission(container, "task", "read"), workflowHandler.GetWorkflowInstance)
		workflowRoutes.POST("/instances/:instance_id/cancel", middleware.RequirePermission(container, "task", "approve"), workflowHandler.CancelWorkflow)
		workflowRoutes.GET("/instances/:instance_id/history", middleware.RequirePermission(container, "task", "read"), workflowHandler.GetWorkflowHistory)
	}

	// 技能管理路由
	skills := authenticated.Group("/skills")
	{
		skills.GET("", middleware.RequirePermission(container, "skill", "read"), skillHandler.ListSkills)
		skills.POST("", middleware.RequirePermission(container, "skill", "create"), skillHandler.CreateSkill)
		skills.GET("/:id", middleware.RequirePermission(container, "skill", "read"), skillHandler.GetSkill)
		skills.PUT("/:id", middleware.RequirePermission(container, "skill", "update"), skillHandler.UpdateSkill)
		skills.DELETE("/:id", middleware.RequirePermission(container, "skill", "delete"), skillHandler.DeleteSkill)
		skills.GET("/categories", middleware.RequirePermission(container, "skill", "read"), skillHandler.GetSkillCategories)
		skills.POST("/assign", middleware.RequirePermission(container, "skill", "update"), skillHandler.AssignSkillToEmployee)
		skills.DELETE("/employees/:employee_id/skills/:skill_id", middleware.RequirePermission(container, "skill", "update"), skillHandler.RemoveSkillFromEmployee)
		skills.GET("/employees/:employee_id", middleware.RequirePermission(container, "skill", "read"), skillHandler.GetEmployeeSkills)
	}

	// 部门管理路由
	departments := authenticated.Group("/departments")
	{
		departments.GET("", middleware.RequirePermission(container, "department", "read"), departmentHandler.ListDepartments)
		departments.POST("", middleware.RequirePermission(container, "department", "create"), departmentHandler.CreateDepartment)
		departments.GET("/:id", middleware.RequirePermission(container, "department", "read"), departmentHandler.GetDepartment)
		departments.PUT("/:id", middleware.RequirePermission(container, "department", "update"), departmentHandler.UpdateDepartment)
		departments.DELETE("/:id", middleware.RequirePermission(container, "department", "delete"), departmentHandler.DeleteDepartment)
		departments.GET("/tree", middleware.RequirePermission(container, "department", "read"), departmentHandler.GetDepartmentTree)
		departments.GET("/roots", middleware.RequirePermission(container, "department", "read"), departmentHandler.GetRootDepartments)
		departments.GET("/:id/sub", middleware.RequirePermission(container, "department", "read"), departmentHandler.GetSubDepartments)
		departments.PUT("/:id/manager", middleware.RequirePermission(container, "department", "update"), departmentHandler.UpdateDepartmentManager)
	}

	// 职位管理路由
	positions := authenticated.Group("/positions")
	{
		positions.GET("", middleware.RequirePermission(container, "position", "read"), positionHandler.ListPositions)
		positions.POST("", middleware.RequirePermission(container, "position", "create"), positionHandler.CreatePosition)
		positions.GET("/:id", middleware.RequirePermission(container, "position", "read"), positionHandler.GetPosition)
		positions.PUT("/:id", middleware.RequirePermission(container, "position", "update"), positionHandler.UpdatePosition)
		positions.DELETE("/:id", middleware.RequirePermission(container, "position", "delete"), positionHandler.DeletePosition)
		positions.GET("/categories", middleware.RequirePermission(container, "position", "read"), positionHandler.GetPositionCategories)
		positions.GET("/category/:category", middleware.RequirePermission(container, "position", "read"), positionHandler.GetPositionsByCategory)
		positions.GET("/level/:level", middleware.RequirePermission(container, "position", "read"), positionHandler.GetPositionsByLevel)
	}

	// 项目管理路由
	projectRoutes := v1.Group("/projects")
	projectRoutes.Use(middleware.Auth(container))
	{
		projectRoutes.POST("", middleware.RequirePermission(container, "project", "create"), projectHandler.CreateProject)
		projectRoutes.GET("", middleware.RequirePermission(container, "project", "read"), projectHandler.ListProjects)
		projectRoutes.GET("/:id", middleware.RequirePermission(container, "project", "read"), projectHandler.GetProject)
		projectRoutes.PUT("/:id", middleware.RequirePermission(container, "project", "update"), projectHandler.UpdateProject)
		projectRoutes.DELETE("/:id", middleware.RequirePermission(container, "project", "delete"), projectHandler.DeleteProject)
		projectRoutes.GET("/:id/members", middleware.RequirePermission(container, "project", "read"), projectHandler.GetProjectMembers)
		projectRoutes.POST("/:id/members", middleware.RequirePermission(container, "project", "update"), projectHandler.AddProjectMember)
		projectRoutes.DELETE("/:id/members/:member_id", middleware.RequirePermission(container, "project", "update"), projectHandler.RemoveProjectMember)
		projectRoutes.GET("/status/:status", middleware.RequirePermission(container, "project", "read"), projectHandler.GetProjectsByStatus)
	}

	// 入职工作流路由
	onboardingHandler := handlers.NewOnboardingHandler(container.GetServiceManager().OnboardingService(), container.GetLogger())
	onboardingRoutes := v1.Group("/onboarding")
	onboardingRoutes.Use(middleware.Auth(container))
	{
		// HR操作：创建待入职员工
		onboardingRoutes.POST("/pending", middleware.RequirePermission(container, "employee", "create"), onboardingHandler.CreatePendingEmployee)
		
		// 部门经理操作：确认入职
		onboardingRoutes.POST("/confirm", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.ConfirmOnboarding)
		
		// HR/管理员操作：完成入职手续
		onboardingRoutes.POST("/:employee_id/probation", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.CompleteProbation)
		
		// 管理员操作：试用期转正
		onboardingRoutes.POST("/confirm-employee", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.ConfirmEmployee)
		
		// 管理员操作：状态变更
		onboardingRoutes.POST("/change-status", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.ChangeEmployeeStatus)
		
		// 查询操作
		onboardingRoutes.GET("/workflows", middleware.RequirePermission(container, "employee", "read"), onboardingHandler.GetOnboardingWorkflows)
		onboardingRoutes.GET("/:employee_id/history", middleware.RequirePermission(container, "employee", "read"), onboardingHandler.GetOnboardingHistory)
		
		// 入职审批工作流操作
		approvalRoutes := onboardingRoutes.Group("/approval")
		{
			// 启动入职审批流程
			approvalRoutes.POST("/start", middleware.RequirePermission(container, "employee", "create"), onboardingHandler.StartOnboardingApproval)
			
			// 处理入职审批决策
			approvalRoutes.POST("/process", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.ProcessOnboardingApproval)
			
			// 获取待审批的入职申请
			approvalRoutes.GET("/pending", middleware.RequirePermission(container, "employee", "read"), onboardingHandler.GetPendingOnboardingApprovals)
			
			// 获取入职审批历史
			approvalRoutes.GET("/history/:employee_id", middleware.RequirePermission(container, "employee", "read"), onboardingHandler.GetOnboardingApprovalHistory)
			
			// 取消入职审批流程
			approvalRoutes.POST("/cancel/:instance_id", middleware.RequirePermission(container, "employee", "update"), onboardingHandler.CancelOnboardingApproval)
		}
	}

	// 权限分配路由
	permissionRoutes := v1.Group("/permissions")
	permissionRoutes.Use(middleware.Auth(container))
	{
		// 权限模板管理
		templateRoutes := permissionRoutes.Group("/templates")
		{
			templateRoutes.POST("", middleware.RequirePermission(container, "permission", "create"), permissionAssignmentHandler.CreatePermissionTemplate)
			templateRoutes.GET("", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.ListPermissionTemplates)
			templateRoutes.GET("/:id", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.GetPermissionTemplate)
			templateRoutes.PUT("/:id", middleware.RequirePermission(container, "permission", "update"), permissionAssignmentHandler.UpdatePermissionTemplate)
			templateRoutes.DELETE("/:id", middleware.RequirePermission(container, "permission", "delete"), permissionAssignmentHandler.DeletePermissionTemplate)
		}

		// 权限分配管理
		assignmentRoutes := permissionRoutes.Group("/assignments")
		{
			assignmentRoutes.POST("", middleware.RequirePermission(container, "permission", "create"), permissionAssignmentHandler.AssignPermissions)
			assignmentRoutes.GET("", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.ListPermissionAssignments)
			assignmentRoutes.GET("/:id", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.GetPermissionAssignment)
			assignmentRoutes.DELETE("/:id", middleware.RequirePermission(container, "permission", "update"), permissionAssignmentHandler.RevokePermissionAssignment)
			assignmentRoutes.GET("/history", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.GetPermissionAssignmentHistory)
		}

		// 入职权限配置管理
		onboardingConfigRoutes := permissionRoutes.Group("/onboarding-configs")
		{
			onboardingConfigRoutes.POST("", middleware.RequirePermission(container, "permission", "create"), permissionAssignmentHandler.CreateOnboardingPermissionConfig)
			onboardingConfigRoutes.GET("", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.ListOnboardingPermissionConfigs)
			onboardingConfigRoutes.GET("/:id", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.GetOnboardingPermissionConfig)
			onboardingConfigRoutes.PUT("/:id", middleware.RequirePermission(container, "permission", "update"), permissionAssignmentHandler.UpdateOnboardingPermissionConfig)
			onboardingConfigRoutes.DELETE("/:id", middleware.RequirePermission(container, "permission", "delete"), permissionAssignmentHandler.DeleteOnboardingPermissionConfig)
		}

		// 权限审批管理
		approvalRoutes := permissionRoutes.Group("/approvals")
		{
			approvalRoutes.GET("/pending", middleware.RequirePermission(container, "permission", "read"), permissionAssignmentHandler.GetPendingPermissionApprovals)
			approvalRoutes.POST("/:id/process", middleware.RequirePermission(container, "permission", "update"), permissionAssignmentHandler.ProcessPermissionApproval)
		}
	}
}
