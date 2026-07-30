package main

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServerLeavesWriteMarginAfterAuthOperationTimeout(t *testing.T) {
	t.Parallel()

	const operationTimeout = 37 * time.Second
	server := newHTTPServer(":8080", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), operationTimeout)

	if got := server.WriteTimeout - operationTimeout; got < 3*time.Second {
		t.Fatalf("write timeout margin = %v, want at least 3s", got)
	}
}
