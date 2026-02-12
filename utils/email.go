package utils

import (
	"fmt"
	"os"

	"github.com/go-gomail/gomail"
)

// SendEmail sends an email using SMTP
func SendEmail(to, subject, body string) error {
	// Check if in development mode
	env := os.Getenv("APP_ENV") // "development" or "production"
	if env == "development" {
		// Just print to console instead of sending
		fmt.Printf("=== Sending Email ===\nTo: %s\nSubject: %s\nBody:\n%s\n==================\n", to, subject, body)
		return nil
	}

	// SMTP credentials from environment
	smtpHost := os.Getenv("SMTP_HOST")     // e.g. smtp.gmail.com
	smtpPort := os.Getenv("SMTP_PORT")     // e.g. 587
	smtpUser := os.Getenv("SMTP_USER")     // your SMTP username
	smtpPass := os.Getenv("SMTP_PASSWORD") // your SMTP password or app password

	port := 587 // default
	fmt.Sscanf(smtpPort, "%d", &port)

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	fmt.Printf("Email sent to %s successfully\n", to)
	return nil
}
