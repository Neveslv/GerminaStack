package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Timeout     time.Duration
}

type SMTPMailer struct {
	config SMTPConfig
}

func NewSMTPMailer(config SMTPConfig) (*SMTPMailer, error) {
	if config.Host == "" ||
		config.Port < 1 ||
		config.Port > 65535 ||
		config.Username == "" ||
		config.Password == "" ||
		config.FromAddress == "" ||
		config.FromName == "" ||
		config.Timeout <= 0 {
		return nil, errors.New("invalid SMTP configuration")
	}
	if strings.ContainsAny(config.Host+config.Username+config.FromAddress+config.FromName, "\r\n") {
		return nil, errors.New("invalid SMTP configuration")
	}
	parsed, err := mail.ParseAddress(config.FromAddress)
	if err != nil || parsed.Address != config.FromAddress {
		return nil, errors.New("invalid SMTP configuration")
	}
	return &SMTPMailer{config: config}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := m.wireMessage(message)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(m.config.Host, strconv.Itoa(m.config.Port))
	dialer := net.Dialer{Timeout: m.config.Timeout}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer connection.Close()

	deadline := time.Now().Add(m.config.Timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set SMTP deadline: %w", err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: m.config.Host,
	}
	var client *smtp.Client
	if m.config.Port == 465 {
		tlsConnection := tls.Client(connection, tlsConfig)
		if err := tlsConnection.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("negotiate SMTP TLS: %w", err)
		}
		client, err = smtp.NewClient(tlsConnection, m.config.Host)
	} else {
		client, err = smtp.NewClient(connection, m.config.Host)
		if err == nil {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				_ = client.Close()
				return errors.New("SMTP server does not support STARTTLS")
			}
			err = client.StartTLS(tlsConfig)
		}
	}
	if err != nil {
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer client.Close()

	authenticator := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	if err := client.Auth(authenticator); err != nil {
		return fmt.Errorf("authenticate SMTP: %w", err)
	}
	if err := client.Mail(m.config.FromAddress); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(message.To); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := writer.Write(raw); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP client: %w", err)
	}
	return nil
}

func (m *SMTPMailer) wireMessage(message Message) ([]byte, error) {
	if strings.ContainsAny(message.To+message.Subject, "\r\n") {
		return nil, errors.New("invalid mail headers")
	}
	recipient, err := mail.ParseAddress(message.To)
	if err != nil || recipient.Address != message.To {
		return nil, errors.New("invalid mail recipient")
	}

	from := (&mail.Address{Name: m.config.FromName, Address: m.config.FromAddress}).String()
	subject := mime.QEncoding.Encode("UTF-8", message.Subject)
	var builder strings.Builder
	builder.WriteString("From: ")
	builder.WriteString(from)
	builder.WriteString("\r\nTo: ")
	builder.WriteString(message.To)
	builder.WriteString("\r\nSubject: ")
	builder.WriteString(subject)
	builder.WriteString("\r\nMIME-Version: 1.0")
	builder.WriteString("\r\nContent-Type: text/plain; charset=UTF-8")
	builder.WriteString("\r\nContent-Transfer-Encoding: 8bit")
	builder.WriteString("\r\n\r\n")

	body := strings.ReplaceAll(message.Body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\n", "\r\n")
	builder.WriteString(body)
	if !strings.HasSuffix(body, "\r\n") {
		builder.WriteString("\r\n")
	}
	return []byte(builder.String()), nil
}

var _ MailSender = (*SMTPMailer)(nil)
