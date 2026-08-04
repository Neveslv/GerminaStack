package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
)

func TestUserHandlerRegisterCreatesSafeInstitutionalUser(t *testing.T) {
	t.Parallel()
	repository := &userRepositoryFake{user: database.User{
		ID: 42, YearID: 7, Name: "Ana Silva", Username: "ana.silva", Email: "ana.silva@institutojef.org.br", IsAdmin: false,
	}}
	handler := NewUserHandler(repository, 2*time.Second)
	recorder := performRequest(handler.Register, http.MethodPost, "/api/users", `{"email":"ana.silva@institutojef.org.br","password":"password-123","password_confirmation":"password-123","year_id":7}`, "application/json")

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.registration.YearID != 7 || repository.registration.Name != "Ana Silva" || repository.registration.Username != "ana.silva" || repository.registration.Email != "ana.silva@institutojef.org.br" {
		t.Fatalf("registration = %#v", repository.registration)
	}
	if repository.registration.PasswordHash == "password-123" || auth.CheckPassword(repository.registration.PasswordHash, "password-123") != nil {
		t.Fatal("registration password was not bcrypt hashed")
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON: %v", err)
	}
	if len(response) != 6 || response["email"] != "ana.silva@institutojef.org.br" || response["is_admin"] != false {
		t.Fatalf("unsafe or incomplete response = %#v", response)
	}
	if _, exists := response["password"]; exists {
		t.Fatalf("response exposes password: %#v", response)
	}
}

func TestUserHandlerRegisterRejectsInvalidInputBeforePersistence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
	}{
		{name: "uppercase email", body: `{"email":"Ana.Silva@institutojef.org.br","password":"password-123","password_confirmation":"password-123","year_id":7}`},
		{name: "other domain", body: `{"email":"ana.silva@example.com","password":"password-123","password_confirmation":"password-123","year_id":7}`},
		{name: "short password", body: `{"email":"ana.silva@institutojef.org.br","password":"1234567","password_confirmation":"1234567","year_id":7}`},
		{name: "long password bytes", body: `{"email":"ana.silva@institutojef.org.br","password":"` + strings.Repeat("\u00e9", 37) + `","password_confirmation":"` + strings.Repeat("\u00e9", 37) + `","year_id":7}`},
		{name: "mismatch", body: `{"email":"ana.silva@institutojef.org.br","password":"password-123","password_confirmation":"password-456","year_id":7}`},
		{name: "nonpositive year", body: `{"email":"ana.silva@institutojef.org.br","password":"password-123","password_confirmation":"password-123","year_id":0}`},
		{name: "caller admin", body: `{"email":"ana.silva@institutojef.org.br","password":"password-123","password_confirmation":"password-123","year_id":7,"is_admin":true}`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := &userRepositoryFake{}
			recorder := performRequest(NewUserHandler(repository, time.Second).Register, http.MethodPost, "/api/users", tt.body, "application/json")
			if recorder.Code != http.StatusBadRequest || repository.calls != 0 {
				t.Fatalf("status/calls = %d/%d, want 400/0; body=%s", recorder.Code, repository.calls, recorder.Body.String())
			}
		})
	}
}

func TestUserHandlerRegisterAcceptsSeventyTwoPasswordBytes(t *testing.T) {
	t.Parallel()
	password := strings.Repeat("\u00e9", 36)
	repository := &userRepositoryFake{user: database.User{ID: 42, YearID: 7, Name: "Ana Silva", Username: "ana.silva", Email: "ana.silva@institutojef.org.br"}}
	body := `{"email":"ana.silva@institutojef.org.br","password":"` + password + `","password_confirmation":"` + password + `","year_id":7}`
	recorder := performRequest(NewUserHandler(repository, 2*time.Second).Register, http.MethodPost, "/api/users", body, "application/json")
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestUserHandlerRegisterMapsRepositoryErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "duplicate", err: database.ErrCredentialConflict, status: http.StatusConflict},
		{name: "unknown year", err: database.ErrYearNotFound, status: http.StatusNotFound},
		{name: "outage", err: errors.New("private database detail"), status: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := &userRepositoryFake{err: tt.err}
			recorder := performRequest(NewUserHandler(repository, 2*time.Second).Register, http.MethodPost, "/api/users", `{"email":"ana.silva@institutojef.org.br","password":"password-123","password_confirmation":"password-123","year_id":7}`, "application/json")
			if recorder.Code != tt.status || strings.Contains(recorder.Body.String(), "private") {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.status, recorder.Body.String())
			}
		})
	}
}

type userRepositoryFake struct {
	registration database.UserRegistration
	user         database.User
	err          error
	calls        int
}

func (f *userRepositoryFake) CreateUser(_ context.Context, registration database.UserRegistration) (database.User, error) {
	f.calls++
	f.registration = registration
	return f.user, f.err
}
