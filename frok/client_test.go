package frok

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientReplyUsesGroqResponsesWithBrowserSearch(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("request = %s / %s", request.Method, request.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(request.Body)
		if err != nil || !strings.Contains(string(body), `"model":"openai/gpt-oss-20b"`) || !strings.Contains(string(body), `"type":"browser_search"`) {
			t.Fatalf("body = %s; err = %v", body, err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"output":[{"content":[{"type":"output_text","text":"Resposta útil"}]}]}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "openai/gpt-oss-20b", time.Second)
	client.endpoint = server.URL
	reply, err := client.Reply(context.Background(), "@frok explique banco de dados")
	if err != nil || reply != "Resposta útil" {
		t.Fatalf("Reply() = %q, %v", reply, err)
	}
}

func TestClientReplyRejectsEmptyOutput(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"output":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "openai/gpt-oss-20b", time.Second)
	client.endpoint = server.URL
	if _, err := client.Reply(context.Background(), "@frok"); err == nil {
		t.Fatal("Reply() error = nil, want empty response error")
	}
}
