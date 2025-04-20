package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	echo     *echo.Echo
	Provider *di.Provider
	logger   *zap.Logger
	config   *config.ServerConfig
}

func NewServer(provider *di.Provider, baseLogger *zap.Logger, configs ...*config.ServerConfig) *Server {
	cfg := setServerConfig(configs...)

	e := echo.New()
	e.HideBanner = cfg.HideBanner
	e.HidePort = cfg.HidePort

	return &Server{
		echo:     e,
		Provider: provider,
		logger:   logger.Named(baseLogger, "Server"),
		config:   cfg,
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.echo.Start(":" + s.config.Port); err != http.ErrServerClosed {
			s.logger.Fatal("fatal error encountered: shutting down server", zap.Error(err))
		}
	}()
	s.logger.Info("server started", zap.String("port", s.config.Port))
}

func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	s.logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown server: %s", zap.Error(err))
	}
	s.logger.Info("server stopped gracefully")
}

func (s *Server) E() *echo.Echo {
	return s.echo
}

func setServerConfig(configs ...*config.ServerConfig) *config.ServerConfig {
	var cfg config.ServerConfig
	if len(configs) > 0 && configs[0] != nil {
		cfg = *configs[0]
	}
	return &cfg
}
