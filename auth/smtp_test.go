package auth

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime/quotedprintable"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestNewSMTPMailerValidatesRequiredConfiguration(t *testing.T) {
	t.Parallel()

	valid := SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "secret",
		FromAddress: "no-reply@example.com",
		FromName:    "GerminaStack",
		Timeout:     5 * time.Second,
	}
	tests := []struct {
		name   string
		mutate func(*SMTPConfig)
	}{
		{name: "host", mutate: func(c *SMTPConfig) { c.Host = "" }},
		{name: "port", mutate: func(c *SMTPConfig) { c.Port = 0 }},
		{name: "username", mutate: func(c *SMTPConfig) { c.Username = "" }},
		{name: "password", mutate: func(c *SMTPConfig) { c.Password = "" }},
		{name: "from address", mutate: func(c *SMTPConfig) { c.FromAddress = "" }},
		{name: "from name", mutate: func(c *SMTPConfig) { c.FromName = "" }},
		{name: "timeout", mutate: func(c *SMTPConfig) { c.Timeout = 0 }},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := valid
			tt.mutate(&cfg)
			if _, err := NewSMTPMailer(cfg); err == nil {
				t.Fatal("NewSMTPMailer() error = nil, want invalid configuration")
			}
		})
	}
}

func TestSMTPMailerBuildsUTF8MIMEMessageWithoutExposingPassword(t *testing.T) {
	t.Parallel()

	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "Equipe GerminaStack",
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}
	raw, err := mailer.wireMessage(Message{
		To:      "ana@example.com",
		Subject: "Seu código de autenticação — GerminaStack",
		Body:    "Código: 123456",
	})
	if err != nil {
		t.Fatalf("wireMessage() error = %v", err)
	}
	text := string(raw)
	for _, want := range []string{
		"From:",
		"To: ana@example.com",
		"Subject:",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("wire message missing %q:\n%s", want, text)
		}
	}
	for _, octet := range raw {
		if octet > 0x7f {
			t.Fatalf("wire message contains non-ASCII octet 0x%x without requiring 8BITMIME", octet)
		}
	}
	decodedBody := decodeQuotedPrintableBody(t, text)
	if !strings.Contains(decodedBody, "Código: 123456") {
		t.Fatalf("decoded body = %q, want authentication code", decodedBody)
	}
	if strings.Contains(text, "smtp-password") {
		t.Fatal("wire message contains SMTP password")
	}
}

func TestSMTPMailerHonorsCanceledContextBeforeDial(t *testing.T) {
	t.Parallel()

	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "GerminaStack",
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = mailer.Send(ctx, Message{To: "ana@example.com", Subject: "subject", Body: "body"})
	if err == nil {
		t.Fatal("Send() error = nil, want canceled context")
	}
}

func TestSMTPMailerSendsQuotedPrintableThroughSTARTTLSWithout8BITMIME(t *testing.T) {
	t.Parallel()

	certificate, roots := smtpTestCertificate(t)
	mailer := newSMTPTransportTestMailer(t, 3*time.Second)
	results := attachSMTPTestServer(t, mailer, func(connection net.Conn) (smtpServerTranscript, error) {
		return serveSTARTTLSSMTP(connection, certificate)
	})
	mailer.tlsConfig = func(serverName string) *tls.Config {
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName,
			RootCAs:    roots,
		}
	}

	err := mailer.Send(context.Background(), Message{
		To:      "ana@example.com",
		Subject: AuthenticationEmailSubject,
		Body:    "Código: 123456",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	result := awaitSMTPServerResult(t, results)
	if result.err != nil {
		t.Fatalf("SMTP test server error = %v", result.err)
	}
	if result.transcript.auth != "\x00mailer\x00smtp-password" {
		t.Fatalf("AUTH payload = %q", result.transcript.auth)
	}
	for _, octet := range []byte(result.transcript.data) {
		if octet > 0x7f {
			t.Fatalf("SMTP DATA contains non-ASCII octet 0x%x without 8BITMIME", octet)
		}
	}
	if !strings.Contains(result.transcript.data, "Content-Transfer-Encoding: quoted-printable") ||
		!strings.Contains(decodeQuotedPrintableBody(t, result.transcript.data), "Código: 123456") ||
		!strings.Contains(result.transcript.data, "Subject:") {
		t.Fatalf("DATA did not contain the message: %q", result.transcript.data)
	}
	if !result.transcript.usedTLS {
		t.Fatal("SMTP session did not upgrade with STARTTLS")
	}
}

func TestSMTPMailerRejectsServerWithoutSTARTTLS(t *testing.T) {
	t.Parallel()

	mailer := newSMTPTransportTestMailer(t, time.Second)
	results := attachSMTPTestServer(t, mailer, serveSMTPWithoutSTARTTLS)

	err := mailer.Send(context.Background(), Message{To: "ana@example.com", Subject: "subject", Body: "body"})
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("Send() error = %v, want missing STARTTLS", err)
	}
	if result := awaitSMTPServerResult(t, results); result.err != nil {
		t.Fatalf("SMTP test server error = %v", result.err)
	}
}

func TestSMTPMailerRejectsInvalidProtocolGreeting(t *testing.T) {
	t.Parallel()

	mailer := newSMTPTransportTestMailer(t, time.Second)
	results := attachSMTPTestServer(t, mailer, serveInvalidSMTPGreeting)

	if err := mailer.Send(context.Background(), Message{To: "ana@example.com", Subject: "subject", Body: "body"}); err == nil {
		t.Fatal("Send() error = nil, want invalid SMTP protocol")
	}
	if result := awaitSMTPServerResult(t, results); result.err != nil {
		t.Fatalf("SMTP test server error = %v", result.err)
	}
}

func TestSMTPMailerHonorsTimeoutDuringProtocol(t *testing.T) {
	t.Parallel()

	const protocolTimeout = 40 * time.Millisecond
	mailer := newSMTPTransportTestMailer(t, protocolTimeout)
	dialogStarted := make(chan struct{})
	results := attachSMTPTestServer(t, mailer, func(connection net.Conn) (smtpServerTranscript, error) {
		return serveBlockedSMTP(connection, dialogStarted)
	})

	startedAt := time.Now()
	err := mailer.Send(context.Background(), Message{To: "ana@example.com", Subject: "subject", Body: "body"})
	elapsed := time.Since(startedAt)
	if err == nil {
		t.Fatal("Send() error = nil, want protocol timeout")
	}
	if elapsed >= time.Second {
		t.Fatalf("Send() elapsed = %v, want prompt configured timeout", elapsed)
	}
	_ = awaitSMTPServerResult(t, results)
}

func TestSMTPMailerHonorsCancellationAfterDial(t *testing.T) {
	t.Parallel()

	const configuredTimeout = 3 * time.Second
	mailer := newSMTPTransportTestMailer(t, configuredTimeout)
	dialogStarted := make(chan struct{})
	results := attachSMTPTestServer(t, mailer, func(connection net.Conn) (smtpServerTranscript, error) {
		return serveBlockedSMTP(connection, dialogStarted)
	})
	ctx, cancel := context.WithCancel(context.Background())
	sendResult := make(chan error, 1)
	go func() {
		sendResult <- mailer.Send(ctx, Message{To: "ana@example.com", Subject: "subject", Body: "body"})
	}()

	select {
	case <-dialogStarted:
	case <-time.After(time.Second):
		t.Fatal("SMTP dialog did not start")
	}
	canceledAt := time.Now()
	cancel()
	select {
	case err := <-sendResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Send() error = %v, want context canceled", err)
		}
		if elapsed := time.Since(canceledAt); elapsed >= 500*time.Millisecond {
			t.Fatalf("cancellation took %v, want less than 500ms", elapsed)
		}
	case <-time.After(time.Second):
		t.Fatalf("Send() did not stop promptly; configured timeout is %v", configuredTimeout)
	}
	_ = awaitSMTPServerResult(t, results)
}

func TestSMTPMailerProductionTLSConfigVerifiesCertificates(t *testing.T) {
	t.Parallel()

	mailer := newSMTPTransportTestMailer(t, time.Second)
	config := mailer.tlsConfig("smtp.example.com")
	if config.ServerName != "smtp.example.com" {
		t.Fatalf("ServerName = %q, want smtp.example.com", config.ServerName)
	}
	if config.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %d, want TLS 1.2", config.MinVersion)
	}
	if config.InsecureSkipVerify {
		t.Fatal("production TLS config enables InsecureSkipVerify")
	}
	if config.RootCAs != nil {
		t.Fatal("production TLS config replaces system roots")
	}
}

type smtpServerTranscript struct {
	auth    string
	data    string
	usedTLS bool
}

type smtpServerResult struct {
	transcript smtpServerTranscript
	err        error
}

func newSMTPTransportTestMailer(t *testing.T, timeout time.Duration) *SMTPMailer {
	t.Helper()
	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:        "smtp.test",
		Port:        587,
		Username:    "mailer",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "GerminaStack",
		Timeout:     timeout,
	})
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}
	return mailer
}

func attachSMTPTestServer(t *testing.T, mailer *SMTPMailer, server func(net.Conn) (smtpServerTranscript, error)) <-chan smtpServerResult {
	t.Helper()
	results := make(chan smtpServerResult, 1)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen(): %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		serverConnection, err := listener.Accept()
		if err != nil {
			results <- smtpServerResult{err: err}
			return
		}
		transcript, err := server(serverConnection)
		results <- smtpServerResult{transcript: transcript, err: err}
	}()
	dialer := &net.Dialer{}
	mailer.dialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp", listener.Addr().String())
	}
	return results
}

func serveSTARTTLSSMTP(connection net.Conn, certificate tls.Certificate) (smtpServerTranscript, error) {
	defer connection.Close()
	var transcript smtpServerTranscript
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	if err := writeSMTPResponse(writer, "220 smtp.test ESMTP ready\r\n"); err != nil {
		return transcript, err
	}
	if _, err := expectSMTPCommand(reader, "EHLO "); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "250-smtp.test\r\n250 STARTTLS\r\n"); err != nil {
		return transcript, err
	}
	if _, err := expectSMTPCommand(reader, "STARTTLS"); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "220 begin TLS\r\n"); err != nil {
		return transcript, err
	}

	tlsConnection := tls.Server(connection, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	})
	if err := tlsConnection.Handshake(); err != nil {
		return transcript, fmt.Errorf("TLS handshake: %w", err)
	}
	defer tlsConnection.Close()
	transcript.usedTLS = true
	reader = bufio.NewReader(tlsConnection)
	writer = bufio.NewWriter(tlsConnection)

	if _, err := expectSMTPCommand(reader, "EHLO "); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "250-smtp.test\r\n250 AUTH PLAIN\r\n"); err != nil {
		return transcript, err
	}
	authCommand, err := expectSMTPCommand(reader, "AUTH PLAIN ")
	if err != nil {
		return transcript, err
	}
	encodedAuth := strings.TrimPrefix(authCommand, "AUTH PLAIN ")
	decodedAuth, err := base64.StdEncoding.DecodeString(encodedAuth)
	if err != nil {
		return transcript, fmt.Errorf("decode AUTH: %w", err)
	}
	transcript.auth = string(decodedAuth)
	if err := writeSMTPResponse(writer, "235 authentication successful\r\n"); err != nil {
		return transcript, err
	}

	if _, err := expectSMTPCommand(reader, "MAIL FROM:<no-reply@example.com>"); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "250 sender accepted\r\n"); err != nil {
		return transcript, err
	}
	if _, err := expectSMTPCommand(reader, "RCPT TO:<ana@example.com>"); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "250 recipient accepted\r\n"); err != nil {
		return transcript, err
	}
	if _, err := expectSMTPCommand(reader, "DATA"); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "354 send message\r\n"); err != nil {
		return transcript, err
	}
	data, err := textproto.NewReader(reader).ReadDotBytes()
	if err != nil {
		return transcript, fmt.Errorf("read DATA: %w", err)
	}
	transcript.data = string(data)
	if err := writeSMTPResponse(writer, "250 queued\r\n"); err != nil {
		return transcript, err
	}
	if _, err := expectSMTPCommand(reader, "QUIT"); err != nil {
		return transcript, err
	}
	if err := writeSMTPResponse(writer, "221 goodbye\r\n"); err != nil {
		return transcript, err
	}
	return transcript, nil
}

func serveSMTPWithoutSTARTTLS(connection net.Conn) (smtpServerTranscript, error) {
	defer connection.Close()
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	if err := writeSMTPResponse(writer, "220 smtp.test ESMTP ready\r\n"); err != nil {
		return smtpServerTranscript{}, err
	}
	if _, err := expectSMTPCommand(reader, "EHLO "); err != nil {
		return smtpServerTranscript{}, err
	}
	return smtpServerTranscript{}, writeSMTPResponse(writer, "250 smtp.test\r\n")
}

func serveInvalidSMTPGreeting(connection net.Conn) (smtpServerTranscript, error) {
	defer connection.Close()
	writer := bufio.NewWriter(connection)
	return smtpServerTranscript{}, writeSMTPResponse(writer, "invalid SMTP greeting\r\n")
}

func serveBlockedSMTP(connection net.Conn, dialogStarted chan<- struct{}) (smtpServerTranscript, error) {
	defer connection.Close()
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	if err := writeSMTPResponse(writer, "220 smtp.test ESMTP ready\r\n"); err != nil {
		return smtpServerTranscript{}, err
	}
	if _, err := expectSMTPCommand(reader, "EHLO "); err != nil {
		return smtpServerTranscript{}, err
	}
	close(dialogStarted)
	_, err := io.Copy(io.Discard, reader)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return smtpServerTranscript{}, err
	}
	return smtpServerTranscript{}, nil
}

func expectSMTPCommand(reader *bufio.Reader, prefix string) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("SMTP command = %q, want prefix %q", line, prefix)
	}
	return line, nil
}

func writeSMTPResponse(writer *bufio.Writer, response string) error {
	if _, err := writer.WriteString(response); err != nil {
		return err
	}
	return writer.Flush()
}

func awaitSMTPServerResult(t *testing.T, results <-chan smtpServerResult) smtpServerResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(time.Second):
		t.Fatal("SMTP test server did not finish")
		return smtpServerResult{}
	}
}

func decodeQuotedPrintableBody(t *testing.T, wireMessage string) string {
	t.Helper()
	normalized := strings.ReplaceAll(wireMessage, "\r\n", "\n")
	_, encodedBody, found := strings.Cut(normalized, "\n\n")
	if !found {
		t.Fatalf("wire message has no header/body separator: %q", wireMessage)
	}
	decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(encodedBody)))
	if err != nil {
		t.Fatalf("decode quoted-printable body: %v", err)
	}
	return string(decoded)
}

func smtpTestCertificate(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "smtp.test"},
		DNSNames:              []string{"smtp.test"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create test certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse test certificate: %v", err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsed)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: privateKey}, roots
}
