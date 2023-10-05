package main

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain string
	Host string
	Port int
	Username string
	Password string
	Encryption string
	FromAddress string
	FromName string
}

type Message struct {
	From string
	FromName string
	To string
	Subject string
	Attachments []string
	Data any
	DataMap map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}
	msg.DataMap=data

	formateMessage,err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMesssage,err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}
	
	mailServer := mail.NewSMTPClient()
	mailServer.Host = m.Host
	mailServer.Port = m.Port
	mailServer.Username = m.Username
	mailServer.Password = m.Password
	mailServer.Encryption = m.getEncryption(m.Encryption)
	mailServer.KeepAlive = false
	mailServer.ConnectTimeout = 10 * time.Second
	mailServer.SendTimeout = 10 * time.Second

	smtpClient,err := mailServer.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMesssage)
	email.AddAlternative(mail.TextHTML,formateMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string,error) {
	templateToRender := "./templates/mail.html.gohtml"

	t,err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", nil
	}
	
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "",err
	}
	
	formatedMessage := tpl.String()
	formatedMessage,err = m.inlineCSS(formatedMessage)
	if err != nil {
		return "", nil
	}

	return formatedMessage,nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string,error) {
	templateToRender := "./templates/mail.plain.gohtml"

	t,err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", nil
	}
	
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "",err
	}
	
	plainMsg := tpl.String()

	return plainMsg,nil
}

func (m *Mail) inlineCSS(s string) (string,error) {
	options := premailer.Options{
		RemoveClasses: false,
		CssToAttributes: false,
		KeepBangImportant: true,
	} 

	prem,err := premailer.NewPremailerFromString(s,&options)
	if err != nil {
		return "", nil
	}

	html,err :=  prem.Transform()
	if err != nil {
		return "", nil
	}

	return html,err
}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}