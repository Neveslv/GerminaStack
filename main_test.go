package main

import (
	"net/http"
	"testing"
	"time"

	"germinaStack/auth"
)

func TestNewHTTPServerLeavesWriteMarginAfterAuthOperationTimeout(t *testing.T) {
	t.Parallel()

	const operationTimeout = 37 * time.Second
	server := newHTTPServer(":8080", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), operationTimeout)

	minimumMargin := server.ReadTimeout + auth.ChallengeInvalidationTimeout + httpResponseGrace
	if got := server.WriteTimeout - operationTimeout; got < minimumMargin {
		t.Fatalf("write timeout margin = %v, want at least %v", got, minimumMargin)
	}
}
