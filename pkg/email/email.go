package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"net/url"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromName     string
	FromEmail    string
	FrontendURL  string
}

// EmailService handles email sending
type EmailService struct {
	config EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{config: config}
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(toEmail, token string) error {
	// Build the reset URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		s.config.FrontendURL,
		url.QueryEscape(token),
		url.QueryEscape(toEmail),
	)

	// Parse and execute the HTML template
	htmlContent, err := s.renderPasswordResetEmail(toEmail, resetURL)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Build the email
	subject := "Reset Your Password - Investify"
	message := s.buildHTMLEmail(toEmail, subject, htmlContent)

	// Send the email
	return s.sendEmail(toEmail, message)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Gmail requires TLS authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildHTMLEmail builds an HTML email message
func (s *EmailService) buildHTMLEmail(to, subject, htmlBody string) []byte {
	headers := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n",
		s.config.FromName,
		s.config.FromEmail,
		to,
		subject,
	)

	return []byte(headers + htmlBody)
}

// renderPasswordResetEmail renders the password reset email template
func (s *EmailService) renderPasswordResetEmail(email, resetURL string) (string, error) {
	tmpl, err := template.New("password_reset").Parse(passwordResetTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		Email    string
		ResetURL string
		AppName  string
	}{
		Email:    email,
		ResetURL: resetURL,
		AppName:  "Investify",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// passwordResetTemplate is the HTML template for password reset emails
const passwordResetTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7fa;">
    <table role="presentation" style="width: 100%; border-collapse: collapse;">
        <tr>
            <td style="padding: 40px 0;">
                <table role="presentation" style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 30px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px; font-weight: 600;">{{.AppName}}</h1>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #1a1a2e; margin: 0 0 20px 0; font-size: 24px; font-weight: 600;">Reset Your Password</h2>
                            
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                Hello,
                            </p>
                            
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                We received a request to reset the password for the account associated with <strong>{{.Email}}</strong>.
                            </p>
                            
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
                                Click the button below to reset your password. This link will expire in <strong>1 hour</strong>.
                            </p>
                            
                            <!-- CTA Button -->
                            <table role="presentation" style="margin: 0 auto 30px auto;">
                                <tr>
                                    <td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border-radius: 8px;">
                                        <a href="{{.ResetURL}}" style="display: inline-block; padding: 16px 32px; color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600;">
                                            Reset Password
                                        </a>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0 0 20px 0;">
                                If you didn't request this password reset, you can safely ignore this email. Your password will remain unchanged.
                            </p>
                            
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0;">
                                If the button above doesn't work, copy and paste this link into your browser:
                            </p>
                            <p style="color: #667eea; font-size: 14px; line-height: 1.6; margin: 10px 0 0 0; word-break: break-all;">
                                <a href="{{.ResetURL}}" style="color: #667eea;">{{.ResetURL}}</a>
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8fafc; padding: 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                            <p style="color: #a0aec0; font-size: 14px; margin: 0 0 10px 0;">
                                This email was sent by {{.AppName}}
                            </p>
                            <p style="color: #cbd5e0; font-size: 12px; margin: 0;">
                                © 2026 {{.AppName}}. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`
