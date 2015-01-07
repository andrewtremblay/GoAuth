package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"
)

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), "<>")
}

func Send(email, link string) error {

	// Connect to the remote SMTP server.
	c, err := smtp.Dial("127.0.0.1:2525")
	if err != nil {
		return err
	}
	// Set the sender and recipient.
	c.Mail("sender@example.org")
	c.Rcpt(email)
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	buf := bytes.NewBufferString("Please confirm this email " + link)
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}
	return nil

}

func SendWithAuthentication() {

	// Set up authentication information.

	smtpServer := "127.0.0.1:2525"
	auth := smtp.PlainAuth(
		"",
		"admin",
		"admin",
		smtpServer,
	)

	from := mail.Address{"example", "info@example.com"}
	to := mail.Address{"customer", "customer@example.com"}
	title := "Mail"

	body := "This is an email confirmation."

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		smtpServer,
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
		//[]byte("This is the email body."),
	)
	if err != nil {
		log.Fatal(err)
	}
}
