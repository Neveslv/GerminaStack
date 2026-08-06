package config

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"germinaStack/auth"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	TwoFactorSecret      string
	CookieSecure         bool
	HTTPAddr             string
	AuthOperationTimeout time.Duration
	SMTP                 auth.SMTPConfig
	Frok                 FrokConfig
}

type FrokConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
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
	if err := validateAuthenticationSecrets(jwtSecret, twoFactorSecret); err != nil {
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
	authOperationTimeout := 15 * time.Second
	if value := strings.TrimSpace(os.Getenv("AUTH_OPERATION_TIMEOUT")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return Config{}, errors.New("AUTH_OPERATION_TIMEOUT is invalid")
		}
		authOperationTimeout = parsed
	}
	frokTimeout := 30 * time.Second
	if value := strings.TrimSpace(os.Getenv("FROK_TIMEOUT")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return Config{}, errors.New("FROK_TIMEOUT is invalid")
		}
		frokTimeout = parsed
	}
	frokModel := strings.TrimSpace(os.Getenv("FROK_MODEL"))
	if frokModel == "" {
		frokModel = "openai/gpt-oss-20b"
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
		DatabaseURL:          databaseURL,
		JWTSecret:            jwtSecret,
		TwoFactorSecret:      twoFactorSecret,
		CookieSecure:         cookieSecure,
		HTTPAddr:             httpAddr,
		AuthOperationTimeout: authOperationTimeout,
		SMTP:                 smtpConfig,
		Frok: FrokConfig{
			APIKey:  strings.TrimSpace(os.Getenv("GROQ_API_KEY")),
			Model:   frokModel,
			Timeout: frokTimeout,
		},
	}, nil
}

func validateAuthenticationSecrets(jwtSecret, twoFactorSecret string) error {
	if len([]byte(jwtSecret)) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 bytes")
	}
	if len([]byte(twoFactorSecret)) < 32 {
		return errors.New("TWO_FACTOR_SECRET must contain at least 32 bytes")
	}
	jwtDigest := sha256.Sum256([]byte(jwtSecret))
	twoFactorDigest := sha256.Sum256([]byte(twoFactorSecret))
	if subtle.ConstantTimeCompare(jwtDigest[:], twoFactorDigest[:]) == 1 {
		return errors.New("JWT_SECRET and TWO_FACTOR_SECRET must be different")
	}
	return nil
}

func required(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
