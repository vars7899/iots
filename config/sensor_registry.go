package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var SensorSchemaRepository *SensorSchemaConfig

type SensorSchemaConfig struct {
	Schema []SensorSchema `json:"schema" mapstructure:"schema"`
}

type SensorSchema struct {
	SensorCode    string        `json:"sensor_code" mapstructure:"sensor_code"`
	SchemaVersion int           `json:"schema_version" mapstructure:"schema_version"`
	Fields        []SensorField `json:"fields" mapstructure:"fields"`
}

type SensorField struct {
	Code     int64  `json:"code" mapstructure:"code"`
	Name     string `json:"name" mapstructure:"name"`
	Unit     string `json:"unit" mapstructure:"unit"`
	DataType string `json:"data_type" mapstructure:"data_type"`
}

type SensorSchemaRegistry struct {
	config *SensorSchemaConfig
	mu     sync.RWMutex
	v      *viper.Viper
	l      *zap.Logger
}

func LoadSensorSchemaRegistry(filename string, filetype string, path string, baseLogger *zap.Logger) error {
	l := logger.Named(baseLogger, "SensorSchemaRegistry")
	completePath := fmt.Sprintf("%s/%s.%s", path, filename, filetype)

	v := viper.New()
	v.SetConfigName(filename)
	v.SetConfigType(filetype)
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		l.Error("error reading configuration", zap.String("path", completePath), zap.Error(err))
		return apperror.ErrConfigLoad.WithMessage("error while reading sensor schema config").Wrap(err)
	}

	schemaCfg := new(SensorSchemaConfig)
	if err := v.Unmarshal(&schemaCfg); err != nil {
		l.Error("failed to unmarshal sensor schema config", zap.String("path", completePath), zap.Error(err))
		return apperror.ErrConfigParse.WithMessage("error while parsing sensor schema config").Wrap(err)
	}

	reg := &SensorSchemaRegistry{
		config: schemaCfg,
		v:      v,
		l:      l,
	}

	fmt.Println(schemaCfg)

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		l.Info("Sensor schema config file changed. Reloading...", zap.String("event", e.Name))
		var updated SensorSchemaConfig
		if err := v.Unmarshal(&updated); err != nil {
			l.Error("Failed to reload sensor schema config", zap.Error(err))
			return
		}

		reg.mu.Lock()
		reg.config = &updated
		reg.mu.Unlock()

		l.Info("[Reload]: Sensor schema config reloaded successfully.")
	})

	SensorSchemaRepository = reg.config
	l.Info("Sensor schema config loaded successfully.")

	return nil
}

func (r *SensorSchemaRegistry) GetConfig() *SensorSchemaConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}
