package dto

import "github.com/vars7899/iots/internal/validatorz"

type RegisterUserRequestDTO struct {
	Email       string `json:"email" validate:"required,email,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	UserName    string `json:"username" validate:"required,alphanum,min=3,max=100"`
	Password    string `json:"password" validate:"required,min=8,max=100"`
}

func (dto RegisterUserRequestDTO) Validate() error { return validatorz.Validate.Struct(dto) }

type LoginUserRequestDTO struct {
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	UserName    string `json:"username,omitempty" validate:"omitempty,min=3"`
	Password    string `json:"password" validate:"required,min=8"`
}

func (dto LoginUserRequestDTO) Validate() error { return validatorz.Validate.Struct(dto) }
