package v1

import (
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/auth/token"
)

type APIDependencies struct {
	SensorService *service.SensorService
	UserService   *service.UserService
	TokenService  token.TokenService
}
