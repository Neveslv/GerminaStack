package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

func performAccountRequestWithAuth(handler gin.HandlerFunc, method, path, body string, isAdmin, authenticated bool) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		if authenticated {
			c.Set(auth.ContextUserID, int64(42))
			c.Set(auth.ContextIsAdmin, isAdmin)
		}
		handler(c)
	})
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

var _ = http.MethodGet
