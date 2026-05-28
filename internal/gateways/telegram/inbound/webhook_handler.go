package inbound

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot/models"
)

const telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token" //nolint:gosec // G101: header name, not a credential

type WebhookHandler struct {
	handler *LazyHandler
	log     *slog.Logger
}

func NewWebhookHandler(handler *LazyHandler, log *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		handler: handler,
		log:     log.With("component", "webhook_handler"),
	}
}

func (wh *WebhookHandler) Handle(c *gin.Context) {
	var update models.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		wh.log.ErrorContext(c.Request.Context(), "failed to decode update", "err", err)
		c.AbortWithStatus(http.StatusBadRequest)

		return
	}

	bgCtx := context.WithoutCancel(c.Request.Context())

	go func() {
		wh.handler.Handle(bgCtx, nil, &update)
	}()

	c.Status(http.StatusOK)
}

func (wh *WebhookHandler) SecretMiddleware(secret string) gin.HandlerFunc {
	expected := []byte(secret)

	return func(c *gin.Context) {
		got := []byte(c.GetHeader(telegramSecretHeader))
		if subtle.ConstantTimeCompare(got, expected) != 1 {
			wh.log.WarnContext(c.Request.Context(), "secret mismatch", "remote_addr", c.ClientIP())
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		c.Next()
	}
}
