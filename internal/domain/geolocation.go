package domain

type GeoLocation struct {
	Latitude         float64  `gorm:"type:double precision;not null;index:idx_location" json:"latitude"`
	Longitude        float64  `gorm:"type:double precision;not null;index:idx_location" json:"longitude"`
	Accuracy         *float64 `gorm:"type:double precision" json:"accuracy,omitempty"`
	Altitude         *float64 `gorm:"type:double precision" json:"altitude,omitempty"`
	AltitudeAccuracy *float64 `gorm:"type:double precision" json:"altitude_accuracy,omitempty"`
	Description      string   `gorm:"type:text;default:''" json:"description,omitempty"`
}
