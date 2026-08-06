package auth

import (
	"errors"
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

const AuthenticationEmailSubject = "Seu c\u00f3digo de autentica\u00e7\u00e3o \u2014 GerminaStack"

var sixDigitCode = regexp.MustCompile(`^[0-9]{6}$`)

type Message struct { To, Subject, Body string }

func AuthenticationMessage(username, email, code string) (Message, error) {
	if strings.ContainsAny(username, "\r\n") { return Message{}, errors.New("invalid username") }
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || strings.ContainsAny(email, "\r\n") { return Message{}, errors.New("invalid recipient") }
	if !sixDigitCode.MatchString(code) { return Message{}, errors.New("invalid authentication code") }
	body := fmt.Sprintf(`<div style="font-family:Arial,sans-serif;max-width:500px;margin:auto;padding:32px;color:#234"><h2 style="color:#26d">GerminaStack</h2><p>Ol\u00e1, %s!</p><p>Use este c\u00f3digo para confirmar seu login:</p><p style="font-size:32px;font-weight:bold;letter-spacing:8px;text-align:center;background:#f1f;padding:18px;border-radius:10px;color:#26d">%s</p><p style="color:#678">Este c\u00f3digo \u00e9 v\u00e1lido por 10 minutos e pode ser usado uma \u00fanica vez (uso \u00fanico).</p><p style="font-size:13px;color:#678">Se voc\u00ea n\u00e3o reconhece esta tentativa, ignore este e-mail e altere sua senha.</p><p>Equipe GerminaStack</p></div>`, html.EscapeString(username), code)
	for _, codepoint := range []string{"00e1", "00f3", "00e7", "00e3", "00e9", "00fa", "00ed"} { body = strings.ReplaceAll(body, `\u`+codepoint, string(rune(mustUnquote(codepoint)))) }
	return Message{To: email, Subject: AuthenticationEmailSubject, Body: body}, nil
}

func mustUnquote(code string) rune { value, _ := strconv.Unquote(`"\u` + code + `"`); return []rune(value)[0] }
