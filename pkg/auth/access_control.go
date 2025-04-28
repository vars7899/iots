package auth

import (
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type accessControlService struct {
	enforcer *casbin.Enforcer
	db       *gorm.DB
	l        *zap.Logger
}

func NewAccessControlService(db *gorm.DB, modelPath string, baseLogger *zap.Logger) (AccessControlService, error) {
	l := logger.Named(baseLogger, "AccessControlService")

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		l.Error("failed to create new gorm db adapter", zap.Error(err))
		return nil, apperror.ErrInit.WithMessage("failed to initialize casbin service").Wrap(err).AsInternal()
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		l.Error("failed to create new casbin enforcer", zap.Error(err))
		return nil, apperror.ErrInit.WithMessage("failed to initialize casbin service").Wrap(err).AsInternal()
	}

	if err := enforcer.LoadPolicy(); err != nil {
		l.Error("failed to load policy from enforcer", zap.Error(err))
		return nil, apperror.ErrLoad.WithMessage("failed to load casbin service policies").Wrap(err).AsInternal()
	}

	return &accessControlService{
		enforcer: enforcer,
		db:       db,
		l:        l,
	}, nil
}

func (s *accessControlService) Enforce(subject string, object string, action string) (bool, error) {
	return s.enforcer.Enforce(subject, object, action)
}

func (s *accessControlService) LoadPolicy() error {
	return s.enforcer.LoadPolicy()
}

func (s *accessControlService) AddRoleForUser(user string, role string) (bool, error) {
	return s.enforcer.AddRoleForUser(user, role)
}

func (s *accessControlService) DeleteRoleForUser(user string, role string) (bool, error) {
	return s.enforcer.DeleteRoleForUser(user, role)
}

func (s *accessControlService) AddPolicy(role string, resource string, action string) (bool, error) {
	return s.enforcer.AddPolicy(role, resource, action)
}

func (s *accessControlService) RemovePolicy(role string, resource string, action string) (bool, error) {
	return s.enforcer.RemovePolicy(role, resource, action)
}

func (s *accessControlService) CheckPermission(userID uuid.UUID, resource string, action string) (bool, error) {
	return s.enforcer.Enforce(userID.String(), resource, action)
}

func (s *accessControlService) SyncUserRoles(user *model.User) error {
	s.enforcer.DeleteRolesForUser(user.ID.String())

	for _, role := range user.Roles {
		if _, err := s.enforcer.AddRoleForUser(user.ID.String(), role.Slug); err != nil {
			return err
		}
	}

	return s.enforcer.SavePolicy()
}

func SplitPermissionCode(code string) []string {
	return strings.Split(code, ":")
}
