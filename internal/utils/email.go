package utils

import (
	"crypto/tls"
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func SendWelcomeEmail(toEmail string, username string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	senderEmail := os.Getenv("SMTP_EMAIL")   // Your Gmail
	senderPassword := os.Getenv("SMTP_PASS") // Your App Password

	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Welcome to Expense Tracker!")
	m.SetBody("text/html", fmt.Sprintf("<h1>Hi %s!</h1><p>Thanks for signing up. Start tracking your expenses today!</p>", username))

	d := gomail.NewDialer(smtpHost, smtpPort, senderEmail, senderPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
