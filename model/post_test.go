package model

import (
	"strings"
	"testing"
)

func TestPostValidateImagePair(t *testing.T) {
	tests := []struct {
		name        string
		imageURL    *string
		description *string
		wantErr     bool
	}{
		{name: "ambos ausentes"},
		{
			name:        "ambos presentes",
			imageURL:    stringPointer("post.png"),
			description: stringPointer("Diagrama"),
		},
		{
			name:     "somente URL",
			imageURL: stringPointer("post.png"),
			wantErr:  true,
		},
		{
			name:        "somente descrição",
			description: stringPointer("Diagrama"),
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
			post := Post{
				ImageURL:         test.imageURL,
				ImageDescription: test.description,
			}

			err := post.Validate()
			if test.wantErr {
				if err == nil {
					t.Fatal("Validate() deveria rejeitar um par de imagem incompleto")
				}
				if !strings.Contains(err.Error(), "imagem do post") ||
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
