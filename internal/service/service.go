package service

import (
	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
)

type CasbinService interface {
	Enforce(subject string, object string, action string) (bool, error)
	LoadPolicy() error
	AddRoleForUser(user string, role string) (bool, error)
	DeleteRoleForUser(user string, role string) (bool, error)
	AddPolicy(role string, resource string, action string) (bool, error)
	RemovePolicy(role string, resource string, action string) (bool, error)
	CheckPermission(userID uuid.UUID, resource string, action string) (bool, error)
	SyncUserRoles(user *model.User) error
}
