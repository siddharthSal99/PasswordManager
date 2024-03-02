package main

import (
	gomail "gopkg.in/mail.v2"
)

func (s *Server) sendCode(code string, email string) error {
	// TODO: Implement email service send code.
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", s.sourceEmailUsername+"@"+s.sourceEmailHost)

	// Set E-Mail receivers
	m.SetHeader("To", email)

	// Set E-Mail subject
	m.SetHeader("Subject", "SSPASSMAN validation code")

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "Your validation code: "+code)

	// Settings for SMTP server
	// fmt.Println("username:", s.sourceEmailUsername, "password:", s.sourceEmailPassword)
	d := gomail.NewDialer("smtp.gmail.com", 587, s.sourceEmailUsername+"@"+s.sourceEmailHost, s.sourceEmailPassword)
	// d.StartTLSPolicy = gomail.MandatoryStartTLS

	// // This is only needed when SSL/TLS certificate is not valid on server.
	// // In production this should be set to false.
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: false,
	// 	ServerName: "smtp.gmail.com"}

	// Now send E-Mail
	return d.DialAndSend(m)
}
