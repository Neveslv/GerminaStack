package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type NotificationRepository interface {
	ListNotifications(context.Context, int64) ([]model.Notification, error)
	MarkNotificationsAsRead(context.Context, int64) error
}

type NotificationHandler struct {
	repository       NotificationRepository
	operationTimeout time.Duration
}

func NewNotificationHandler(repository NotificationRepository, operationTimeout time.Duration) *NotificationHandler {
	return &NotificationHandler{repository: repository, operationTimeout: operationTimeout}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nÃƒÆ’Ã‚Â£o autorizado"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	notifications, err := h.repository.ListNotifications(ctx, userID)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nÃƒÆ’Ã‚Â£o autorizado"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if err := h.repository.MarkNotificationsAsRead(ctx, userID); err != nil {
		writeNotificationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func notificationUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(auth.ContextUserID)
	userID, ok := value.(int64)
	return userID, exists && ok && userID > 0
}

func writeNotificationError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrCredentialNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conta nÃƒÆ’Ã‚Â£o encontrada"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviÃƒÆ’Ã‚Â§o indisponÃƒÆ’Ã‚Â­vel"})
}
