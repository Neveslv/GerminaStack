package handlers

import "github.com/gin-gonic/gin"

// Explicit method names keep the handler API readable for direct composition
// tests while the existing route registrations remain compatible.
func (h *ReactionHandler) ToggleReaction(c *gin.Context) {
	h.Toggle(c)
}

func (h *ReactionHandler) GetReaction(c *gin.Context) {
	h.Get(c)
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	h.List(c)
}

func (h *NotificationHandler) MarkNotificationsAsRead(c *gin.Context) {
	h.MarkRead(c)
}
