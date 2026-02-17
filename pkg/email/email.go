package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"net/url"
	"time"
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

// LowStockProduct holds details of a product that has reached low stock
type LowStockProduct struct {
	Name          string
	Code          string
	Quantity      int
	QuantityAlert int
}

// SendLowStockAlertEmail sends a low stock alert email to an admin
func (s *EmailService) SendLowStockAlertEmail(toEmail, orgName string, products []LowStockProduct) error {
	htmlContent, err := s.renderLowStockAlertEmail(orgName, products)
	if err != nil {
		return fmt.Errorf("failed to render low stock alert email: %w", err)
	}

	subject := fmt.Sprintf("⚠️ Low Stock Alert - %s", orgName)
	message := s.buildHTMLEmail(toEmail, subject, htmlContent)

	return s.sendEmail(toEmail, message)
}

// renderLowStockAlertEmail renders the low stock alert email template
func (s *EmailService) renderLowStockAlertEmail(orgName string, products []LowStockProduct) (string, error) {
	tmpl, err := template.New("low_stock_alert").Parse(lowStockAlertTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		OrgName  string
		Products []LowStockProduct
		AppName  string
		Year     int
	}{
		OrgName:  orgName,
		Products: products,
		AppName:  "Investify",
		Year:     time.Now().Year(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendEmail sends an email using SMTP with timeout protection
func (s *EmailService) sendEmail(to string, message []byte) error {
	addr := net.JoinHostPort(s.config.SMTPHost, fmt.Sprintf("%d", s.config.SMTPPort))

	// Create a dialer with timeout to prevent indefinite blocking
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	// Connect with timeout
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Set read/write deadlines to prevent hanging during SMTP conversation
	deadline := time.Now().Add(60 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set connection deadline: %w", err)
	}

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Start TLS for Gmail (required for port 587)
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start email data: %w", err)
	}

	if _, err := wc.Write(message); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close email data: %w", err)
	}

	// Quit gracefully
	if err := client.Quit(); err != nil {
		// Log but don't fail - email was already sent
		return nil
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
		Year     int
	}{
		Email:    email,
		ResetURL: resetURL,
		AppName:  "Investify",
		Year:     time.Now().Year(),
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
                                © {{.Year}} {{.AppName}}. All rights reserved.
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

// lowStockAlertTemplate is the HTML template for low stock alert emails
const lowStockAlertTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Low Stock Alert</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7fa;">
    <table role="presentation" style="width: 100%; border-collapse: collapse;">
        <tr>
            <td style="padding: 40px 0;">
                <table role="presentation" style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #e53e3e 0%, #c53030 100%); padding: 40px 30px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px; font-weight: 600;">⚠️ Low Stock Alert</h1>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #1a1a2e; margin: 0 0 20px 0; font-size: 24px; font-weight: 600;">{{.OrgName}}</h2>
                            
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                The following products have reached or fallen below their low stock threshold and may need to be restocked:
                            </p>
                            
                            <!-- Products Table -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 0 0 30px 0;">
                                <thead>
                                    <tr style="background-color: #f7fafc;">
                                        <th style="padding: 12px 16px; text-align: left; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0;">Product</th>
                                        <th style="padding: 12px 16px; text-align: left; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0;">Code</th>
                                        <th style="padding: 12px 16px; text-align: center; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0;">Current Qty</th>
                                        <th style="padding: 12px 16px; text-align: center; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0;">Alert At</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{range .Products}}
                                    <tr>
                                        <td style="padding: 12px 16px; font-size: 14px; color: #2d3748; border-bottom: 1px solid #e2e8f0;">{{.Name}}</td>
                                        <td style="padding: 12px 16px; font-size: 14px; color: #718096; border-bottom: 1px solid #e2e8f0;">{{.Code}}</td>
                                        <td style="padding: 12px 16px; font-size: 14px; color: #e53e3e; font-weight: 600; text-align: center; border-bottom: 1px solid #e2e8f0;">{{.Quantity}}</td>
                                        <td style="padding: 12px 16px; font-size: 14px; color: #718096; text-align: center; border-bottom: 1px solid #e2e8f0;">{{.QuantityAlert}}</td>
                                    </tr>
                                    {{end}}
                                </tbody>
                            </table>
                            
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0;">
                                Please restock these products at your earliest convenience to avoid running out of inventory.
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
                                © {{.Year}} {{.AppName}}. All rights reserved.
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
