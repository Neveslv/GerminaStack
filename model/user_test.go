package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func stringPointer(value string) *string {
	return &value
}

func pointerTo[T any](value T) *T {
	return &value
}

func TestUserValidateProfileImagePair(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    *string
		description *string
		wantErr     bool
	}{
		{name: "ambos ausentes"},
		{
			name:        "ambos presentes",
			imageURL:    stringPointer("profile.png"),
			description: stringPointer("Foto de perfil"),
		},
		{
			name:     "somente URL",
			imageURL: stringPointer("profile.png"),
			wantErr:  true,
		},
		{
			name:        "somente descrição",
			description: stringPointer("Foto de perfil"),
			wantErr:     true,
		},
		{
			name:        "ambos presentes e vazios",
			imageURL:    stringPointer(""),
			description: stringPointer(""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user := User{
				ProfileImageURL:         test.imageURL,
				ProfileImageDescription: test.description,
			}

			err := user.Validate()
			if test.wantErr {
				if err == nil {
					t.Fatal("Validate() deveria rejeitar um par de imagem incompleto")
				}
				if !strings.Contains(err.Error(), "imagem de perfil") ||
					!strings.Contains(err.Error(), "descrição") {
					t.Fatalf("Validate() retornou erro pouco claro: %q", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, esperado nil", err)
			}
		})
	}
}

func TestUserJSONUsesSnakeCaseAndHidesPassword(t *testing.T) {
	const secret = "valor-ultrassecreto-9f53"

	user := User{
		YearID:                  pointerTo(int64(7)),
		ProfileImageURL:         stringPointer("profile.png"),
		ProfileImageDescription: stringPointer("Foto de perfil"),
		Username:                "aluna",
		Password:                secret,
	}

	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := decoded["password"]; exists {
		t.Fatal("password não pode aparecer no JSON")
	}
	if strings.Contains(string(payload), secret) {
		t.Fatal("o valor secreto de password não pode aparecer no JSON")
	}
	if got := decoded["id_year"]; got != float64(7) {
		t.Fatalf("id_year = %v, esperado 7", got)
	}
	if got := decoded["profile_image_url"]; got != "profile.png" {
		t.Fatalf("profile_image_url = %v, esperado profile.png", got)
	}
	if got := decoded["profile_image_description"]; got != "Foto de perfil" {
		t.Fatalf(
			"profile_image_description = %v, esperado Foto de perfil",
			got,
		)
	}
	if _, exists := decoded["profileImageURL"]; exists {
		t.Fatal("profileImageURL em camelCase não pode aparecer no JSON")
	}
}
