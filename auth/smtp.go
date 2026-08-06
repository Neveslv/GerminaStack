package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
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
	config      SMTPConfig
	dialContext func(context.Context, string, string) (net.Conn, error)
	tlsConfig   func(string) *tls.Config
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
	dialer := &net.Dialer{Timeout: config.Timeout}
	return &SMTPMailer{
		config:      config,
		dialContext: dialer.DialContext,
		tlsConfig: func(serverName string) *tls.Config {
			return &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: serverName,
			}
		},
	}, nil
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
	connection, err := m.dialContext(ctx, "tcp", address)
	if err != nil {
		return smtpContextError(ctx, "connect SMTP", err)
	}
	defer connection.Close()

	deadline := time.Now().Add(m.config.Timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return smtpContextError(ctx, "set SMTP deadline", err)
	}
	stopCancellation := context.AfterFunc(ctx, func() {
		_ = connection.SetDeadline(time.Now())
	})
	defer stopCancellation()

	tlsConfig := m.tlsConfig(m.config.Host)
	var client *smtp.Client
	if m.config.Port == 465 {
		tlsConnection := tls.Client(connection, tlsConfig)
		if err := tlsConnection.HandshakeContext(ctx); err != nil {
			return smtpContextError(ctx, "negotiate SMTP TLS", err)
		}
		client, err = smtp.NewClient(tlsConnection, m.config.Host)
	} else {
		client, err = smtp.NewClient(connection, m.config.Host)
		if err == nil {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				_ = client.Close()
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}
				return errors.New("SMTP server does not support STARTTLS")
			}
			err = client.StartTLS(tlsConfig)
		}
	}
	if err != nil {
		return smtpContextError(ctx, "initialize SMTP client", err)
	}
	defer client.Close()

	authenticator := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	if err := client.Auth(authenticator); err != nil {
		return smtpContextError(ctx, "authenticate SMTP", err)
	}
	if err := client.Mail(m.config.FromAddress); err != nil {
		return smtpContextError(ctx, "set SMTP sender", err)
	}
	if err := client.Rcpt(message.To); err != nil {
		return smtpContextError(ctx, "set SMTP recipient", err)
	}
	writer, err := client.Data()
	if err != nil {
		return smtpContextError(ctx, "start SMTP message", err)
	}
	if _, err := writer.Write(raw); err != nil {
		_ = writer.Close()
		return smtpContextError(ctx, "write SMTP message", err)
	}
	if err := writer.Close(); err != nil {
		return smtpContextError(ctx, "finish SMTP message", err)
	}
	if err := client.Quit(); err != nil {
		return smtpContextError(ctx, "quit SMTP client", err)
	}
	return nil
}

func smtpContextError(ctx context.Context, operation string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return fmt.Errorf("%s: %w", operation, err)
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
	builder.WriteString("\r\nContent-Transfer-Encoding: quoted-printable")
	builder.WriteString("\r\n\r\n")

	body := strings.ReplaceAll(message.Body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	quotedWriter := quotedprintable.NewWriter(&builder)
	if _, err := quotedWriter.Write([]byte(body)); err != nil {
		return nil, fmt.Errorf("encode mail body: %w", err)
	}
	if err := quotedWriter.Close(); err != nil {
		return nil, fmt.Errorf("finish mail body encoding: %w", err)
	}
	return []byte(builder.String()), nil
}

var _ MailSender = (*SMTPMailer)(nil)
