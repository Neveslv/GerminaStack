package handlers

import (
	"context"
	"net/http"
	"time"

	"germinaStack/auth"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type CatalogRepository interface {
	ListYears(context.Context) ([]model.Year, error)
	ListSubjects(context.Context, int64) ([]model.Subject, error)
}

type CatalogHandler struct {
	repository       CatalogRepository
	operationTimeout time.Duration
}

func NewCatalogHandler(repository CatalogRepository, operationTimeout time.Duration) *CatalogHandler {
	return &CatalogHandler{repository: repository, operationTimeout: operationTimeout}
}

type yearResponse struct {
	ID   int64  `json:"id"`
	Year string `json:"year"`
}

type subjectResponse struct {
	ID         int64  `json:"id"`
	YearID     *int64 `json:"year_id"`
	Subject    string `json:"subject"`
	PostsCount int64  `json:"posts_count"`
}

func (h *CatalogHandler) ListYears(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()

	years, err := h.repository.ListYears(ctx)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	response := make([]yearResponse, 0, len(years))
	for _, year := range years {
		response = append(response, yearDTO(year))
	}
	c.JSON(http.StatusOK, response)
}

func (h *CatalogHandler) ListSubjects(c *gin.Context) {
	value, exists := c.Get(auth.ContextUserID)
	userID, valid := value.(int64)
	if !exists || !valid || userID <= 0 {
		writeCatalogUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	subjects, err := h.repository.ListSubjects(ctx, userID)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	response := make([]subjectResponse, 0, len(subjects))
	for _, subject := range subjects {
		response = append(response, subjectDTO(subject))
	}
	c.JSON(http.StatusOK, response)
}

func writeCatalogError(c *gin.Context, _ error) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
}

func writeCatalogUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
}

func yearDTO(year model.Year) yearResponse {
	return yearResponse{ID: year.ID, Year: year.Year}
}

func subjectDTO(subject model.Subject) subjectResponse {
	return subjectResponse{ID: subject.ID, YearID: subject.YearID, Subject: subject.Subject, PostsCount: subject.PostsCount}
}
