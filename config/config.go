package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"germinaStack/auth"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	TwoFactorSecret string
	CookieSecure    bool
	HTTPAddr        string
	SMTP            auth.SMTPConfig
}

func Load() (Config, error) {
	databaseURL, err := required("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	jwtSecret, err := required("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	twoFactorSecret, err := required("TWO_FACTOR_SECRET")
	if err != nil {
		return Config{}, err
	}
	smtpHost, err := required("SMTP_HOST")
	if err != nil {
		return Config{}, err
	}
	smtpPortText, err := required("SMTP_PORT")
	if err != nil {
		return Config{}, err
	}
	smtpPort, err := strconv.Atoi(smtpPortText)
	if err != nil || smtpPort < 1 || smtpPort > 65535 {
		return Config{}, errors.New("SMTP_PORT is invalid")
	}
	smtpUsername, err := required("SMTP_USERNAME")
	if err != nil {
		return Config{}, err
	}
	smtpPassword, err := required("SMTP_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	smtpFromAddress, err := required("SMTP_FROM_ADDRESS")
	if err != nil {
		return Config{}, err
	}
	smtpFromName, err := required("SMTP_FROM_NAME")
	if err != nil {
		return Config{}, err
	}

	cookieSecure := true
	if value := strings.TrimSpace(os.Getenv("COOKIE_SECURE")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, errors.New("COOKIE_SECURE is invalid")
		}
		cookieSecure = parsed
	}
	httpAddr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	smtpConfig := auth.SMTPConfig{
		Host:        smtpHost,
		Port:        smtpPort,
		Username:    smtpUsername,
		Password:    smtpPassword,
		FromAddress: smtpFromAddress,
		FromName:    smtpFromName,
		Timeout:     10 * time.Second,
	}
	if _, err := auth.NewSMTPMailer(smtpConfig); err != nil {
		return Config{}, errors.New("SMTP configuration is invalid")
	}

	return Config{
		DatabaseURL:     databaseURL,
		JWTSecret:       jwtSecret,
		TwoFactorSecret: twoFactorSecret,
		CookieSecure:    cookieSecure,
		HTTPAddr:        httpAddr,
		SMTP:            smtpConfig,
	}, nil
}

func required(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
