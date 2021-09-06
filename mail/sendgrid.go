package mail

import (
	"errors"
	"fmt"
	"os"

	"sterna/log"

	sendgrid "github.com/sendgrid/sendgrid-go"
	sgMail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailConf interface {
	NewMailServ(l log.Logger) MailServ
}

type MailServ interface {
	Subject(s string) MailServ
	PlaintText(p string) MailServ
	Html(h string) MailServ
	SendSingle(name, mail string) error
}

type SendGridConf struct {
	From *sgMail.Email      `yaml:"from"`
	Bcc  *sgMail.BccSetting `yaml:"bcc"`
}

func (conf *SendGridConf) NewMailServ(l log.Logger) MailServ {
	return &sendGridServ{
		SendGridConf: conf,
		l:            l,
	}
}

type sendGridServ struct {
	*SendGridConf
	l log.Logger

	subject, txt, html string
}

const sendgridKey = "SENDGRID_API_KEY"

func (sgc *sendGridServ) Subject(s string) MailServ {
	sgc.subject = s
	return sgc
}

func (sgc *sendGridServ) PlaintText(p string) MailServ {
	sgc.txt = p
	return sgc
}

func (sgc *sendGridServ) Html(h string) MailServ {
	sgc.html = h
	return sgc
}

func (sg *sendGridServ) SendSingle(name, mail string) error {
	if sg.SendGridConf == nil {
		return errors.New("sendGridConf is nil")
	}
	to := sgMail.NewEmail(name, mail)
	message := sgMail.NewSingleEmail(sg.From, sg.subject, to, sg.txt, sg.html)
	if sg.Bcc != nil && *sg.Bcc.Enable {
		message.SetMailSettings(sgMail.NewMailSettings().SetBCC(sg.Bcc))
	}

	key := os.Getenv(sendgridKey)
	if key == "" {
		return errors.New("missing env key: " + sendgridKey)
	}

	client := sendgrid.NewSendClient(key)
	res, err := client.Send(message)
	if err != nil {
		return err
	} else {
		if res.StatusCode > 299 {
			return fmt.Errorf("status code [%d] mail error: %s", res.StatusCode, res.Body)
		} else {
			sg.l.Info(res.Body)
		}
	}
	return nil
}
