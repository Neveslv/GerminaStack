package auth

import (
	"errors"
	"regexp"
	"strings"
)

var institutionalEmailPattern = regexp.MustCompile(`^([a-z]+)\.([a-z]+)@institutojef\.org\.br$`)

var ErrInvalidInstitutionalEmail = errors.New("e-mail institucional inválido")

type InstitutionalIdentity struct {
	Name     string
	Username string
}

func ParseInstitutionalEmail(email string) (InstitutionalIdentity, error) {
	matches := institutionalEmailPattern.FindStringSubmatch(email)
	if matches == nil {
		return InstitutionalIdentity{}, ErrInvalidInstitutionalEmail
	}

	firstName := matches[1]
	lastName := matches[2]
	return InstitutionalIdentity{
		Name:     titleASCII(firstName) + " " + titleASCII(lastName),
		Username: firstName + "." + lastName,
	}, nil
}

func titleASCII(value string) string {
	return strings.ToUpper(value[:1]) + value[1:]
}
