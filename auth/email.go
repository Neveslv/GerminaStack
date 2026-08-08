package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

const AuthenticationEmailSubject = "Seu código de autenticação — GerminaStack"

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
	body := fmt.Sprintf("Olá, %s!\n\nSeu código de autenticação é:\n\n    %s\n\nEste código é válido por 10 minutos e pode ser usado uma única vez (uso único).\n\nSe você não reconhece esta tentativa, ignore este e-mail e altere sua senha.\n\nEquipe GerminaStack\n", username, code)
	return Message{To: email, Subject: AuthenticationEmailSubject, Body: body}, nil
}
