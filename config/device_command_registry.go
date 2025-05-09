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

type DeviceCommandSchemaConfig struct {
	Commands map[string]CommandSchema `json:"command" mapstructure:"command"`
}

type CommandSchema struct {
	Type          string        `json:"type" mapstructure:"type"`
	Command       string        `json:"command" mapstructure:"command"`
	CommandCode   string        `json:"command_code" mapstructure:"command_code"`
	SchemaVersion int           `json:"schema_version" mapstructure:"schema_version"`
	Direction     string        `json:"direction" mapstructure:"direction"`
	Description   string        `json:"description" mapstructure:"description"`
	Payload       PayloadSchema `json:"payload_schema" mapstructure:"payload_schema"`
}

type PayloadSchema struct {
	Required   []string                           `json:"required" mapstructure:"required"`
	Properties map[string]PayloadPropertiesSchema `json:"properties" mapstructure:"properties"`
}

type PayloadPropertiesSchema struct {
	Name     string `json:"name" mapstructure:"name"`
	DataType string `json:"data_type" mapstructure:"data_type"`
}

type DeviceCommandSchemaRegistry struct {
	config *DeviceCommandSchemaConfig
	mu     sync.RWMutex
	v      *viper.Viper
	l      *zap.Logger
}

func NewDeviceCommandSchemaRegistry() *DeviceCommandSchemaRegistry {
	return &DeviceCommandSchemaRegistry{}
}

func (r *DeviceCommandSchemaRegistry) LoadDeviceCommandSchemaRegistry(filename, filetype, path string, baseLogger *zap.Logger) error {
	l := logger.Named(baseLogger, "DeviceCommandSchemaRegistry")
	completePath := fmt.Sprintf("%s/%s.%s", path, filename, filetype)

	v := viper.New()
	v.SetConfigName(filename)
	v.SetConfigType(filetype)
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		l.Error("error reading configuration", zap.String("path", completePath), zap.Error(err))
		return apperror.ErrConfigLoad.WithMessage("error while reading device command schema config").Wrap(err)
	}

	schemaCfg := new(DeviceCommandSchemaConfig)
	if err := v.Unmarshal(&schemaCfg); err != nil {
		l.Error("failed to unmarshal device command schema config", zap.String("path", completePath), zap.Error(err))
		return apperror.ErrConfigParse.WithMessage("error while parsing device command schema config").Wrap(err)
	}

	r.mu.Lock()
	r.config = schemaCfg
	r.mu.Unlock()

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		l.Info("Device command schema config file changed. Reloading...", zap.String("event", e.Name))
		var updated DeviceCommandSchemaConfig
		if err := v.Unmarshal(&updated); err != nil {
			l.Error("Failed to reload device command schema config", zap.Error(err))
			return
		}

		r.mu.Lock()
		r.config = &updated
		r.mu.Unlock()

		l.Info("[Reload]: Device command schema config reloaded successfully.")
	})

	r.v = v
	r.l = l

	l.Info("Device command schema config loaded successfully.")
	return nil
}

func (r *DeviceCommandSchemaRegistry) GetConfig() *DeviceCommandSchemaConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

func (r *DeviceCommandSchemaRegistry) GetCommandByCode(commandCode string) *CommandSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cmd, ok := r.config.Commands[commandCode]; ok {
		return &cmd
	}
	return nil
}
