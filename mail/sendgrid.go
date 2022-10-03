package mail

import (
	"errors"
	"fmt"

	"github.com/94peter/sterna/log"

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
	SendMulti(tos []*To) error
}

type SendGridConf struct {
	ApiKey string             `yaml:"apiKey"`
	From   *sgMail.Email      `yaml:"from"`
	Bcc    *sgMail.BccSetting `yaml:"bcc"`
}

func (conf *SendGridConf) NewMailServ(l log.Logger) MailServ {
	return &sendGridServ{
		SendGridConf: conf,
		key:          conf.ApiKey,
		l:            l,
	}
}

type sendGridServ struct {
	*SendGridConf
	l   log.Logger
	key string

	subject, txt, html string
}

const ENV_SendgridKey = "SENDGRID_API_KEY"

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

	if sg.key == "" {
		return errors.New("missing env key: " + sg.key)
	}

	client := sendgrid.NewSendClient(sg.key)
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

type To struct {
	Name string
	Mail string
}

func (sg *sendGridServ) SendMulti(tos []*To) error {
	v3mail := sgMail.NewV3Mail()
	v3mail.SetFrom(sg.From)
	v3mail.Subject = sg.subject

	p := sgMail.NewPersonalization()
	for _, t := range tos {
		p.AddTos(&sgMail.Email{
			Name:    t.Name,
			Address: t.Mail,
		})
	}
	v3mail.AddPersonalizations(p)
	var contents []*sgMail.Content
	if sg.txt != "" {
		contents = append(contents, sgMail.NewContent("text/plain", sg.txt))
	}
	if sg.html != "" {
		contents = append(contents, sgMail.NewContent("text/html", sg.html))
	}
	v3mail.AddContent(contents...)
	if len(contents) == 0 {
		return errors.New("no content")
	}
	if sg.Bcc != nil {
		v3mail.SetMailSettings(sgMail.NewMailSettings().SetBCC(sg.Bcc))
	}
	client := sendgrid.NewSendClient(sg.key)

	res, err := client.Send(v3mail)
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
