package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/domain/sensor"
	"gorm.io/gorm"
)

func Seeder() {

}

func SeedSensorData(db *gorm.DB, count int) error {
	if count < 0 {
		return fmt.Errorf("invalid seed argument, count should be a +ve integer")
	}
	for i := 0; i < count; i++ {
		s := model.Sensor{
			DeviceID: gofakeit.UUID(),
			Name:     gofakeit.Name(),
			Type: model.SensorType(gofakeit.RandomString([]string{
				string(sensor.TemperatureSensor),
				string(sensor.HumiditySensor),
				string(sensor.MotionSensor),
			})),
			Unit:      gofakeit.RandomString([]string{"°C", "%", "m/s"}),
			Precision: gofakeit.Number(0, 2),
			Location:  gofakeit.City(),
		}
		if err := db.WithContext(context.Background()).Create(s).Error; err != nil {
			log.Printf("failed to create sensor: %v", err)
		}
	}
	return nil
}
