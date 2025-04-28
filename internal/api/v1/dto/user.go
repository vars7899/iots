package dto

import (
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/validation"
)

type RegisterUserRequestDTO struct {
	Email       string `json:"email" validate:"required,email,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	UserName    string `json:"username" validate:"required,alphanum,min=3,max=100"`
	Password    string `json:"password" validate:"required,min=8,max=100"`
}

func (dto RegisterUserRequestDTO) Validate() error { return validation.Validate.Struct(dto) }

func (dto RegisterUserRequestDTO) AsModel() *model.User {
	return &model.User{
		Email:       dto.Email,
		PhoneNumber: dto.PhoneNumber,
		Username:    dto.UserName,
		Password:    dto.Password,
	}
}

type LoginCredentials struct {
	Email       string
	Username    string
	PhoneNumber string
	Password    string
}

type LoginUserRequestDTO struct {
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	UserName    string `json:"username,omitempty" validate:"omitempty,min=3"`
	Password    string `json:"password" validate:"required,min=8"`
}

func (dto LoginUserRequestDTO) Validate() error { return validation.Validate.Struct(dto) }

func (dto LoginUserRequestDTO) AsModel() *LoginCredentials {
	return &LoginCredentials{
		Email:       dto.Email,
		Username:    dto.UserName,
		PhoneNumber: dto.PhoneNumber,
		Password:    dto.Password,
	}
}

type RequestPasswordResetDTO struct {
	Email string `json:"email" validate:"required,email,max=100"`
}

func (dto RequestPasswordResetDTO) Validate() error { return validation.Validate.Struct(dto) }

type ResetPasswordDTO struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

func (dto ResetPasswordDTO) Validate() error { return validation.Validate.Struct(dto) }
