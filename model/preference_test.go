package model

import (
	"strings"
	"testing"
)

func validPreference() Preference {
	return Preference{
		ContrastTheme: ContrastThemeNormal,
		FontFamily:    FontFamilyNormal,
		FontSpacing:   FontSpacingNormal,
		FontSize:      FontSizeNormal,
	}
}

func TestPreferenceValidate(t *testing.T) {
	tests := []struct {
		name         string
		change       func(*Preference)
		errorContext string
	}{
		{
			name:   "válida",
			change: func(*Preference) {},
		},
		{
			name: "tema inválido",
			change: func(preference *Preference) {
				preference.ContrastTheme = ContrastTheme("inexistente")
			},
			errorContext: "tema de contraste",
		},
		{
			name: "família inválida",
			change: func(preference *Preference) {
				preference.FontFamily = FontFamily("inexistente")
			},
			errorContext: "família de fonte",
		},
		{
			name: "espaçamento inválido",
			change: func(preference *Preference) {
				preference.FontSpacing = FontSpacing("inexistente")
			},
			errorContext: "espaçamento de fonte",
		},
		{
			name: "tamanho inválido",
			change: func(preference *Preference) {
				preference.FontSize = FontSize("inexistente")
			},
			errorContext: "tamanho de fonte",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preference := validPreference()
			test.change(&preference)

			err := preference.Validate()
			if test.errorContext == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, esperado nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() deveria rejeitar %s", test.errorContext)
			}
			if !strings.Contains(err.Error(), test.errorContext) ||
				!strings.Contains(err.Error(), "inexistente") {
				t.Fatalf("Validate() retornou erro pouco claro: %q", err)
			}
		})
	}
}

func TestPreferenceEnumsRecognizeSQLValues(t *testing.T) {
	tests := []struct {
		name          string
		isValid       func(string) bool
		validValues   []string
		invalidValues []string
	}{
		{
			name: "contrast_theme",
			isValid: func(value string) bool {
				return ContrastTheme(value).IsValid()
			},
			validValues: []string{
				"normal",
				"dark",
				"high_contrast",
				"black_yellow",
				"yellow_black",
			},
			invalidValues: []string{"", "inexistente"},
		},
		{
			name: "font_family",
			isValid: func(value string) bool {
				return FontFamily(value).IsValid()
			},
			validValues: []string{
				"normal",
				"arial",
				"verdana",
				"lexend",
				"atkinson_hyperlegible",
				"open_dyslexic",
			},
			invalidValues: []string{"", "inexistente"},
		},
		{
			name: "font_spacing",
			isValid: func(value string) bool {
				return FontSpacing(value).IsValid()
			},
			validValues:   []string{"normal", "pequeno", "grande"},
			invalidValues: []string{"", "inexistente"},
		},
		{
			name: "font_size",
			isValid: func(value string) bool {
				return FontSize(value).IsValid()
			},
			validValues:   []string{"normal", "pequeno", "grande"},
			invalidValues: []string{"", "inexistente"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, value := range test.validValues {
				if !test.isValid(value) {
					t.Errorf("%q deveria ser válido", value)
				}
			}
			for _, value := range test.invalidValues {
				if test.isValid(value) {
					t.Errorf("%q deveria ser inválido", value)
				}
			}
		})
	}
}
