package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string       `gorm:"unique;not null" json:"email"`
	PasswordHash string       `gorm:"not null" json:"-"`
	RoleID       uuid.UUID    `gorm:"type:uuid" json:"role_id"`
	Role         Role         `gorm:"foreignKey:RoleID" json:"role"`
	DepartmentID *uuid.UUID   `gorm:"type:uuid" json:"department_id"`
	Department   *Department  `gorm:"foreignKey:DepartmentID" json:"department"`
	ManagerID    *uuid.UUID   `gorm:"type:uuid" json:"manager_id"`
	Manager      *User        `gorm:"foreignKey:ManagerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"manager,omitempty"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	Nationality  string       `json:"nationality"`
	IsActive     bool         `gorm:"default:true" json:"is_active"`
	Permissions  []Permission `gorm:"many2many:user_permissions" json:"permissions"` // individual overrides
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	RefreshToken string
	RefreshExp   time.Time

	ResetToken string
	ResetExp   time.Time
}

// Helper: check if user has a permission
func (u *User) HasPermission(permissionName string) bool {
	for _, p := range u.Role.Permissions {
		if p.Name == permissionName {
			return true
		}
	}
	for _, p := range u.Permissions {
		if p.Name == permissionName {
			return true
		}
	}
	return false
}

// Helper: check if user has role
func (u *User) HasRole(roleName string) bool {
	return u.Role.Name == roleName
}

func (u User) Name() string {
	return u.FirstName + " " + u.LastName
}
