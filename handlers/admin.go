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

type AdminRepository interface {
	ListUsers(context.Context, database.AdminFilter) ([]model.User, int, error)
	ListPosts(context.Context, database.AdminFilter) ([]database.AdminPost, int, error)
	GetUser(context.Context, int64) (model.User, error)
	SetBanned(context.Context, int64, bool) error
	SetAdmin(context.Context, int64, bool) error
	DeletePost(context.Context, int64) error
}

type AdminHandler struct {
	repository       AdminRepository
	operationTimeout time.Duration
}

func NewAdminHandler(repository AdminRepository, operationTimeout time.Duration) *AdminHandler {
	return &AdminHandler{repository: repository, operationTimeout: operationTimeout}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if _, ok := h.actor(ctx, c); !ok {
		return
	}
	filter, ok := adminFilter(c)
	if !ok {
		return
	}
	users, total, err := h.repository.ListUsers(ctx, filter)
	if err != nil {
		writeAdminError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": total})
}

func (h *AdminHandler) ListPosts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if _, ok := h.actor(ctx, c); !ok {
		return
	}
	filter, ok := adminFilter(c)
	if !ok {
		return
	}
	posts, total, err := h.repository.ListPosts(ctx, filter)
	if err != nil {
		writeAdminError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": posts, "total": total})
}

func (h *AdminHandler) BanUser(c *gin.Context)  { h.changeUser(c, false) }
func (h *AdminHandler) SetAdmin(c *gin.Context) { h.changeUser(c, true) }

func (h *AdminHandler) DeletePost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if _, ok := h.actor(ctx, c); !ok {
		return
	}
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	if err := h.repository.DeletePost(ctx, id); err != nil {
		writeAdminError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) changeUser(c *gin.Context, adminChange bool) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if !decodeAdminInput(c, &input) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	actor, ok := h.actor(ctx, c)
	if !ok {
		return
	}
	target, err := h.repository.GetUser(ctx, id)
	if err != nil {
		writeAdminError(c, err)
		return
	}
	if actor.ID == target.ID || isSuperAdmin(target.Username) || (!isSuperAdmin(actor.Username) && target.IsAdmin) || (adminChange && !isSuperAdmin(actor.Username)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "acesso proibido"})
		return
	}
	if adminChange {
		err = h.repository.SetAdmin(ctx, id, input.Enabled)
	} else {
		err = h.repository.SetBanned(ctx, id, input.Enabled)
	}
	if err != nil {
		writeAdminError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) actor(ctx context.Context, c *gin.Context) (model.User, bool) {
	actorID, ok := c.Get(auth.ContextUserID)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return model.User{}, false
	}
	actor, err := h.repository.GetUser(ctx, actorID.(int64))
	if err != nil {
		writeAdminError(c, err)
		return model.User{}, false
	}
	if !actor.IsAdmin && !isSuperAdmin(actor.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "acesso proibido"})
		return model.User{}, false
	}
	return actor, true
}

func decodeAdminInput(c *gin.Context, target any) bool {
	if err := decodeJSON(c, target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return false
	}
	return true
}
func adminFilter(c *gin.Context) (database.AdminFilter, bool) {
	search := c.Query("q")
	if len(search) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return database.AdminFilter{}, false
	}
	pagination, err := database.ParsePagination(c.Query("page"), "5")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return database.AdminFilter{}, false
	}
	return database.AdminFilter{Search: search, Pagination: pagination}, true
}
func isSuperAdmin(username string) bool {
	return username == "nicolas.oliveira" || username == "matheus.fazan"
}
func writeAdminError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrAdminUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "não foi possível concluir a ação"})
}
