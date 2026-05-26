package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"noirbot/internal/gateways/telegram/inbound"
	"noirbot/pkg/config"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	srv             *http.Server
	shutdownTimeout time.Duration
	log             *slog.Logger
}

func New(cfg *config.Config, engine *gin.Engine, log *slog.Logger) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         *cfg.HTTP.Addr,
			Handler:      engine,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
		},
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
		log:             log.With("component", "http_server"),
	}
}

func (s *Server) Start(_ context.Context) error {
	s.log.Info("starting http server", "addr", s.srv.Addr)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("http server crashed", "err", err)
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(shutdownCtx)
}

func NewEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	return engine
}

func RegisterRoutes(engine *gin.Engine, webhook *inbound.WebhookHandler, cfg *config.Config) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	if cfg.Telegram.WebhookSecret != "" {
		engine.POST("/webhook",
			webhook.SecretMiddleware(cfg.Telegram.WebhookSecret),
			webhook.Handle,
		)
		return
	}
	engine.POST("/webhook", webhook.Handle)
}
