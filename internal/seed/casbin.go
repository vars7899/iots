package seed

import (
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var permissions = []model.Permission{
	// User management
	{Code: "users:read", Name: "Read Users"},
	{Code: "users:create", Name: "Create Users"},
	{Code: "users:update", Name: "Update Users"},
	{Code: "users:delete", Name: "Delete Users"},

	// Sensor management
	{Code: "sensors:read", Name: "Read Sensors"},
	{Code: "sensors:create", Name: "Create Sensors"},
	{Code: "sensors:update", Name: "Update Sensors"},
	{Code: "sensors:delete", Name: "Delete Sensors"},
	{Code: "sensors:configure", Name: "Configure Sensors"},
	{Code: "sensors:calibrate", Name: "Calibrate Sensors"},
	{Code: "sensors:assign", Name: "Assign Sensors to User/Location"},

	// Device management
	{Code: "devices:read", Name: "Read Devices"},
	{Code: "devices:create", Name: "Create Devices"},
	{Code: "devices:update", Name: "Update Devices"},
	{Code: "devices:delete", Name: "Delete Devices"},
	{Code: "devices:restart", Name: "Restart Devices"},
	{Code: "devices:firmware:update", Name: "Update Device Firmware"},

	// Location or site management
	{Code: "locations:read", Name: "Read Locations"},
	{Code: "locations:create", Name: "Create Locations"},
	{Code: "locations:update", Name: "Update Locations"},
	{Code: "locations:delete", Name: "Delete Locations"},

	// Analytics or logs
	{Code: "logs:view", Name: "View Logs"},
	{Code: "logs:export", Name: "Export Logs"},
	{Code: "analytics:view", Name: "View Analytics"},

	// Roles and permissions
	{Code: "roles:read", Name: "Read Roles"},
	{Code: "roles:create", Name: "Create Roles"},
	{Code: "roles:update", Name: "Update Roles"},
	{Code: "roles:delete", Name: "Delete Roles"},

	{Code: "permissions:read", Name: "Read Permissions"},
	{Code: "permissions:assign", Name: "Assign Permissions to Roles"},
}

var roles = []model.Role{
	{
		Slug:        "admin",
		Name:        "Administrator",
		Description: "Full system access",
		IsProtected: true,
	},
	{
		Slug:        "viewer",
		Name:        "Viewer",
		Description: "Read-only access to the system",
		IsProtected: true,
	},
	{
		Slug:        "sensor.read",
		Name:        "Sensor Reader",
		Description: "Read-only access to sensors",
		IsProtected: false,
	},
	{
		Slug:        "sensor.write",
		Name:        "Sensor Writer",
		Description: "Can create and update sensors",
		IsProtected: false,
	},
}

var rolePermissions = map[string][]string{
	"admin": {
		"users:read", "users:create", "users:update", "users:delete",
		"sensors:read", "sensors:create", "sensors:update", "sensors:delete", "sensors:configure",
	},
	"viewer": {
		"users:read", "sensors:read",
	},
	"sensor.read": {
		"sensors:read",
	},
	"sensor.write": {
		"sensors:read", "sensors:create", "sensors:update",
	},
}

func SeedRolesAndPermission(db *gorm.DB, casbinService service.CasbinService, logger *zap.Logger) error {
	logger.Info("Seeding roles and permissions")

	for _, permission := range permissions {
		if err := db.Where(model.Permission{Code: permission.Code}).FirstOrCreate(&permission).Error; err != nil {
			logger.Error("Failed to create permission", zap.String("code", permission.Code), zap.Error(err))
			return err
		}
	}

	for _, role := range roles {
		if err := db.Where(model.Role{Slug: role.Slug}).FirstOrCreate(&role).Error; err != nil {
			logger.Error("Failed to create role", zap.String("slug", role.Slug), zap.Error(err))
			return err
		}
	}

	// Assign permissions to roles in the database
	for roleSlug, permCodes := range rolePermissions {
		// Find the role
		var role model.Role
		if err := db.Where("slug = ?", roleSlug).First(&role).Error; err != nil {
			logger.Error("Failed to find role", zap.String("slug", roleSlug), zap.Error(err))
			continue
		}

		// Find permissions
		var permissions []model.Permission
		if err := db.Where("code IN ?", permCodes).Find(&permissions).Error; err != nil {
			logger.Error("Failed to find permissions", zap.Strings("codes", permCodes), zap.Error(err))
			continue
		}

		// Associate permissions with role
		if err := db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			logger.Error("Failed to associate permissions with role",
				zap.String("role", roleSlug), zap.Error(err))
			continue
		}
	}

	// Seed Casbin policies
	for roleSlug, permCodes := range rolePermissions {
		for _, permCode := range permCodes {
			// Split permission code into resource and action
			// Format: resource:action (e.g., "users:read")
			parts := service.SplitPermissionCode(permCode)
			if len(parts) != 2 {
				logger.Warn("Invalid permission code format", zap.String("code", permCode))
				continue
			}

			resource := parts[0]
			action := parts[1]

			// Add policy to Casbin
			_, err := casbinService.AddPolicy(roleSlug, resource, action)
			if err != nil {
				logger.Error("Failed to add Casbin policy",
					zap.String("role", roleSlug),
					zap.String("resource", resource),
					zap.String("action", action),
					zap.Error(err))
				continue
			}
		}
	}

	// Create a default admin user if none exists
	var count int64
	db.Model(&model.User{}).Count(&count)

	if count == 0 {
		logger.Info("Creating default admin user")

		// Find admin role
		var adminRole model.Role
		if err := db.Where("slug = ?", "admin").First(&adminRole).Error; err != nil {
			logger.Error("Failed to find admin role", zap.Error(err))
			return err
		}

		// Create admin user
		adminUser := model.User{
			Username:    "admin",
			Email:       "admin@example.com",
			PhoneNumber: "+1234567890",
			IsActive:    true,
			Roles:       []model.Role{adminRole},
		}

		// Set password
		if err := adminUser.SetPassword("Admin@123"); err != nil {
			logger.Error("Failed to set admin password", zap.Error(err))
			return err
		}

		// Save admin user
		if err := db.Create(&adminUser).Error; err != nil {
			logger.Error("Failed to create admin user", zap.Error(err))
			return err
		}

		// Sync admin user with Casbin
		if err := casbinService.SyncUserRoles(&adminUser); err != nil {
			logger.Error("Failed to sync admin user roles with Casbin", zap.Error(err))
			return err
		}
	}

	// Save all policies to Casbin
	if err := casbinService.LoadPolicy(); err != nil {
		logger.Error("Failed to load Casbin policies", zap.Error(err))
		return err
	}

	logger.Info("Successfully seeded roles and permissions")
	return nil
}
