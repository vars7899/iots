package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username    string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"`
	Email       string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	Password    string         `gorm:"type:text;not null" json:"password"`
	PhoneNumber string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"phone_number"`
	Roles       []Role         `gorm:"many2many:user_roles;constraints:onUpdate:CASCADE,onDelete:CASCADE;" json:"roles"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastLogin   *time.Time     `json:"last_login"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedBy   *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
}

// TODO: implement public view
type UserPublicView struct {
	Username    string
	Email       string
	PhoneNumber string
	Roles       []Role
	IsActive    bool
}

func (u *User) PublicView() *UserPublicView {
	return &UserPublicView{
		Username:    u.Username,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Roles:       u.Roles,
		IsActive:    u.IsActive,
	}
}

func (u *User) HashPassword(raw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) ComparePassword(inputPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(inputPassword))
}

func (u *User) HasPermission(code string) bool {
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if code == permission.Code {
				return true
			}
		}
	}
	return false
}

type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(50);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permissions;constraints:onUpdate:CASCADE,onDelete:CASCADE;" json:"permissions"`
	IsProtected bool           `gorm:"default:false" json:"is_protected"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (r *Role) CheckIsProtected() error {
	if r.IsProtected {
		return fmt.Errorf("cannot delete protected role")
	}
	return nil
}

type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string    `gorm:"type:varchar(255);unique;not null" json:"code"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Join table for User and Role
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Role      Role      `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
}

// Join table for Role and Permission
type RolePermission struct {
	RoleID       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	Role         Role       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
	Permission   Permission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE"`
}
