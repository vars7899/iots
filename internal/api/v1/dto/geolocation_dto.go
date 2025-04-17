package dto

import (
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/validatorz"
)

type GeoLocationDTO struct {
	Latitude         float64  `json:"latitude" validate:"required,latitude"`
	Longitude        float64  `json:"longitude" validate:"required,longitude"`
	Accuracy         *float64 `json:"accuracy,omitempty" validate:"omitempty,number,gte=0"`
	Altitude         *float64 `json:"altitude,omitempty" validate:"omitempty,number"`
	AltitudeAccuracy *float64 `json:"altitude_accuracy,omitempty" validate:"omitempty,number,gte=0"`
	Description      string   `json:"description,omitempty" validate:"omitempty,max=255"`
}

func (dto *GeoLocationDTO) Validate() error {
	return validatorz.Validate.Struct(&dto)
}

func (dto *GeoLocationDTO) AsModel() *domain.GeoLocation {
	var gl domain.GeoLocation

	gl.Latitude = dto.Latitude
	gl.Longitude = dto.Longitude

	if dto.Accuracy != nil {
		gl.Accuracy = dto.Accuracy
	}
	if dto.Altitude != nil {
		gl.Altitude = dto.Altitude
	}
	if dto.AltitudeAccuracy != nil {
		gl.AltitudeAccuracy = dto.AltitudeAccuracy
	}
	if dto.Description != "" {
		gl.Description = dto.Description
	}
	return &gl
}
