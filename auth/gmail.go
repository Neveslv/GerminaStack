package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GmailMailer struct {
	clientID, clientSecret, refreshToken, from string
	client                                     *http.Client
}

func NewGmailMailer(clientID, clientSecret, refreshToken, from string) (*GmailMailer, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" || from == "" {
		return nil, errors.New("invalid Gmail configuration")
	}
	return &GmailMailer{clientID, clientSecret, refreshToken, from, http.DefaultClient}, nil
}

func (m *GmailMailer) Send(ctx context.Context, message Message) error {
	token, err := m.accessToken(ctx)
	if err != nil {
		return fmt.Errorf("refresh Gmail token: %w", err)
	}
	raw, err := wireGmailMessage(m.from, message)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"raw": base64.RawURLEncoding.EncodeToString(raw)})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Gmail API returned %s", res.Status)
	}
	return nil
}

func (m *GmailMailer) accessToken(ctx context.Context) (string, error) {
	form := url.Values{"client_id": {m.clientID}, "client_secret": {m.clientSecret}, "refresh_token": {m.refreshToken}, "grant_type": {"refresh_token"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return "", fmt.Errorf("token endpoint returned %s", res.Status)
	}
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", errors.New("empty access token")
	}
	return result.AccessToken, nil
}

func wireGmailMessage(from string, message Message) ([]byte, error) {
	mailer, err := NewSMTPMailer(SMTPConfig{FromAddress: from, FromName: "GerminaStack", Timeout: 1})
	if err != nil {
		return nil, err
	}
	return mailer.wireMessage(message)
}
