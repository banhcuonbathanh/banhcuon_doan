package utils
// File: golang/internal/email/smtp_email_service.go


import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPEmailService(config EmailConfig) *SMTPEmailService {
	return &SMTPEmailService{
		host:     config.Host,
		port:     config.Port,
		username: config.Username,
		password: config.Password,
		from:     config.From,
	}
}

func (s *SMTPEmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	subject := "Welcome to Our Platform!"
	body := fmt.Sprintf(`
Dear %s,

Welcome to our platform! We're excited to have you on board.

Your account has been successfully created with email: %s

If you have any questions, please don't hesitate to contact us.

Best regards,
The Team
`, name, email)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
Dear User,

You have requested to reset your password. Please use the following token to reset your password:

Reset Token: %s

This token will expire in 1 hour. If you did not request this password reset, please ignore this email.

Best regards,
The Team
`, resetToken)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailService) SendAccountDeactivationEmail(ctx context.Context, email, name string) error {
	subject := "Account Deactivated"
	body := fmt.Sprintf(`
Dear %s,

Your account has been deactivated as requested.

If this was done in error or if you wish to reactivate your account, please contact our support team.

Best regards,
The Team
`, name)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) error {
	subject := "Email Verification Required"
	body := fmt.Sprintf(`
Dear User,

Please verify your email address by using the following verification token:

Verification Token: %s

This token will expire in 24 hours.

Best regards,
The Team
`, verificationToken)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailService) SendPasswordChangedNotification(ctx context.Context, email, name string) error {
	subject := "Password Changed Successfully"
	body := fmt.Sprintf(`
Dear %s,

Your password has been successfully changed.

If you did not make this change, please contact our support team immediately.

Best regards,
The Team
`, name)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailService) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := []string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}

	message := strings.Join(msg, "\r\n")
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
}

// Alternative implementation for testing or when SMTP is not available
type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	fmt.Printf("Mock: Sending welcome email to %s (%s)\n", name, email)
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) error {
	fmt.Printf("Mock: Sending password reset email to %s with token: %s\n", email, resetToken)
	return nil
}

func (m *MockEmailService) SendAccountDeactivationEmail(ctx context.Context, email, name string) error {
	fmt.Printf("Mock: Sending account deactivation email to %s (%s)\n", name, email)
	return nil
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) error {
	fmt.Printf("Mock: Sending verification email to %s with token: %s\n", email, verificationToken)
	return nil
}

func (m *MockEmailService) SendPasswordChangedNotification(ctx context.Context, email, name string) error {
	fmt.Printf("Mock: Sending password changed notification to %s (%s)\n", name, email)
	return nil
}