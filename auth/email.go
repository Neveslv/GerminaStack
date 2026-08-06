package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

const AuthenticationEmailSubject = "Seu c\u00f3digo de autentica\u00e7\u00e3o \u2014 GerminaStack"

var sixDigitCode = regexp.MustCompile(`^[0-9]{6}$`)

type Message struct {
	To      string
	Subject string
	Body    string
}

func AuthenticationMessage(username, email, code string) (Message, error) {
	if strings.ContainsAny(username, "\r\n") {
		return Message{}, errors.New("invalid username")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || strings.ContainsAny(email, "\r\n") {
		return Message{}, errors.New("invalid recipient")
	}
	if !sixDigitCode.MatchString(code) {
		return Message{}, errors.New("invalid authentication code")
	}
	body := fmt.Sprintf("Ol\u00e1, %s!\n\nSeu c\u00f3digo de autentica\u00e7\u00e3o \u00e9:\n\n    %s\n\nEste c\u00f3digo \u00e9 v\u00e1lido por 10 minutos e pode ser usado uma \u00fanica vez (uso \u00fanico).\n\nSe voc\u00ea n\u00e3o reconhece esta tentativa, ignore este e-mail e altere sua senha.\n\nEquipe GerminaStack\n", username, code)
	return Message{To: email, Subject: AuthenticationEmailSubject, Body: body}, nil
}
