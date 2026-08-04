package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

const maxCatalogLabelBytes = 100

type CatalogRepository interface {
	ListYears(context.Context) ([]model.Year, error)
	CreateYear(context.Context, string) (model.Year, error)
	UpdateYear(context.Context, int64, string) (model.Year, error)
	DeleteYear(context.Context, int64) error
	ListSubjects(context.Context, *int64) ([]model.Subject, error)
	CreateSubject(context.Context, string, *int64) (model.Subject, error)
	UpdateSubject(context.Context, int64, string, *int64) (model.Subject, error)
	DeleteSubject(context.Context, int64) error
}

type CatalogHandler struct {
	repository       CatalogRepository
	operationTimeout time.Duration
}

func NewCatalogHandler(repository CatalogRepository, operationTimeout time.Duration) *CatalogHandler {
	return &CatalogHandler{repository: repository, operationTimeout: operationTimeout}
}

type yearRequest struct {
	Year string `json:"year"`
}

type yearResponse struct {
	ID   int64  `json:"id"`
	Year string `json:"year"`
}

type subjectRequest struct {
	Subject string          `json:"subject"`
	YearID  json.RawMessage `json:"year_id"`
}

type subjectResponse struct {
	ID      int64  `json:"id"`
	YearID  *int64 `json:"year_id"`
	Subject string `json:"subject"`
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
	var yearID *int64
	if values, supplied := c.Request.URL.Query()["year_id"]; supplied {
		if len(values) != 1 {
			writeInvalidCatalogRequest(c)
			return
		}
		parsed, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || parsed <= 0 {
			writeInvalidCatalogRequest(c)
			return
		}
		yearID = &parsed
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	subjects, err := h.repository.ListSubjects(ctx, yearID)
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

func (h *CatalogHandler) CreateYear(c *gin.Context) {
	label, ok := decodeYearRequest(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	year, err := h.repository.CreateYear(ctx, label)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	c.JSON(http.StatusCreated, yearDTO(year))
}

func (h *CatalogHandler) UpdateYear(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	label, ok := decodeYearRequest(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	year, err := h.repository.UpdateYear(ctx, id, label)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	c.JSON(http.StatusOK, yearDTO(year))
}

func (h *CatalogHandler) DeleteYear(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if err := h.repository.DeleteYear(ctx, id); err != nil {
		writeCatalogError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func (h *CatalogHandler) CreateSubject(c *gin.Context) {
	label, yearID, ok := decodeSubjectRequest(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	subject, err := h.repository.CreateSubject(ctx, label, yearID)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	c.JSON(http.StatusCreated, subjectDTO(subject))
}

func (h *CatalogHandler) UpdateSubject(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	label, yearID, ok := decodeSubjectRequest(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	subject, err := h.repository.UpdateSubject(ctx, id, label, yearID)
	if err != nil {
		writeCatalogError(c, err)
		return
	}
	c.JSON(http.StatusOK, subjectDTO(subject))
}

func (h *CatalogHandler) DeleteSubject(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if err := h.repository.DeleteSubject(ctx, id); err != nil {
		writeCatalogError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func decodeYearRequest(c *gin.Context) (string, bool) {
	var request yearRequest
	if err := decodeJSON(c, &request); err != nil {
		writeInvalidCatalogRequest(c)
		return "", false
	}
	label := strings.TrimSpace(request.Year)
	if len(label) == 0 || len([]byte(label)) > maxCatalogLabelBytes {
		writeInvalidCatalogRequest(c)
		return "", false
	}
	return label, true
}

func decodeSubjectRequest(c *gin.Context) (string, *int64, bool) {
	var request subjectRequest
	if err := decodeJSON(c, &request); err != nil {
		writeInvalidCatalogRequest(c)
		return "", nil, false
	}
	label := strings.TrimSpace(request.Subject)
	if len(label) == 0 || len([]byte(label)) > maxCatalogLabelBytes || len(request.YearID) == 0 {
		writeInvalidCatalogRequest(c)
		return "", nil, false
	}
	if string(request.YearID) == "null" {
		return label, nil, true
	}
	var yearID int64
	if err := json.Unmarshal(request.YearID, &yearID); err != nil || yearID <= 0 {
		writeInvalidCatalogRequest(c)
		return "", nil, false
	}
	return label, &yearID, true
}

func positivePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeInvalidCatalogRequest(c)
		return 0, false
	}
	return id, true
}

func writeCatalogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, database.ErrYearNotFound), errors.Is(err, database.ErrSubjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso não encontrado"})
	case errors.Is(err, database.ErrCatalogConflict), errors.Is(err, database.ErrCatalogReferenced):
		c.JSON(http.StatusConflict, gin.H{"error": "recurso em conflito"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
	}
}

func writeInvalidCatalogRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
}
func yearDTO(year model.Year) yearResponse { return yearResponse{ID: year.ID, Year: year.Year} }
func subjectDTO(subject model.Subject) subjectResponse {
	return subjectResponse{ID: subject.ID, YearID: subject.YearID, Subject: subject.Subject}
}
