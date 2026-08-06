package model

import (
	"fmt"
	"time"
)

type ContrastTheme string

const (
	ContrastThemeNormal       ContrastTheme = "normal"
	ContrastThemeDark         ContrastTheme = "dark"
	ContrastThemeHighContrast ContrastTheme = "high_contrast"
	ContrastThemeBlackYellow  ContrastTheme = "black_yellow"
	ContrastThemeYellowBlack  ContrastTheme = "yellow_black"
)

func (theme ContrastTheme) IsValid() bool {
	switch theme {
	case ContrastThemeNormal, ContrastThemeDark, ContrastThemeHighContrast,
		ContrastThemeBlackYellow, ContrastThemeYellowBlack:
		return true
	default:
		return false
	}
}

type FontFamily string

const (
	FontFamilyNormal               FontFamily = "normal"
	FontFamilyArial                FontFamily = "arial"
	FontFamilyVerdana              FontFamily = "verdana"
	FontFamilyLexend               FontFamily = "lexend"
	FontFamilyAtkinsonHyperlegible FontFamily = "atkinson_hyperlegible"
	FontFamilyOpenDyslexic         FontFamily = "open_dyslexic"
)

func (family FontFamily) IsValid() bool {
	switch family {
	case FontFamilyNormal, FontFamilyArial, FontFamilyVerdana,
		FontFamilyLexend, FontFamilyAtkinsonHyperlegible,
		FontFamilyOpenDyslexic:
		return true
	default:
		return false
	}
}

type FontSpacing string

const (
	FontSpacingNormal  FontSpacing = "normal"
	FontSpacingPequeno FontSpacing = "pequeno"
	FontSpacingGrande  FontSpacing = "grande"
)

func (spacing FontSpacing) IsValid() bool {
	switch spacing {
	case FontSpacingNormal, FontSpacingPequeno, FontSpacingGrande:
		return true
	default:
		return false
	}
}

type FontSize string

const (
	FontSizeNormal  FontSize = "normal"
	FontSizePequeno FontSize = "pequeno"
	FontSizeGrande  FontSize = "grande"
)

func (size FontSize) IsValid() bool {
	switch size {
	case FontSizeNormal, FontSizePequeno, FontSizeGrande:
		return true
	default:
		return false
	}
}

type Preference struct {
	ID            int64          `db:"id" json:"id"`
	UserID        int64          `db:"id_user" json:"id_user"`
	ContrastTheme *ContrastTheme `db:"contrast_theme" json:"contrast_theme"`
	FontFamily    *FontFamily    `db:"font_family" json:"font_family"`
	FontSpacing   *FontSpacing   `db:"font_spacing" json:"font_spacing"`
	FontSize      *FontSize      `db:"font_size" json:"font_size"`
	CreatedAt     *time.Time     `db:"created_at" json:"created_at"`
}

func (preference Preference) Validate() error {
	if preference.ContrastTheme != nil && !preference.ContrastTheme.IsValid() {
		return fmt.Errorf("tema de contraste inválido: %q", *preference.ContrastTheme)
	}
	if preference.FontFamily != nil && !preference.FontFamily.IsValid() {
		return fmt.Errorf("família de fonte inválida: %q", *preference.FontFamily)
	}
	if preference.FontSpacing != nil && !preference.FontSpacing.IsValid() {
		return fmt.Errorf("espaçamento de fonte inválido: %q", *preference.FontSpacing)
	}
	if preference.FontSize != nil && !preference.FontSize.IsValid() {
		return fmt.Errorf("tamanho de fonte inválido: %q", *preference.FontSize)
	}

	return nil
}
