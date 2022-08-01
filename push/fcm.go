package push

import (
	"errors"

	"github.com/NaySoftware/go-fcm"
)

type PushDI interface {
	NewPushServ() (PushServ, error)
}

type PushServ interface {
	Push(tokens []string, title, body string, data map[string]interface{}) (err error, errTokens []string)
}

type FcmConf struct {
	ServerKey string   `yaml:"fcmServerKey"`
	Enable    bool     `yaml:"enable"`
	Platform  []string `yaml:"platform"`
}

func (conf *FcmConf) NewPushServ() (PushServ, error) {
	if conf == nil || !conf.Enable {
		return nil, errors.New("not enable push")
	}
	c := fcm.NewFcmClient(conf.ServerKey)
	return &fcmImpl{
		FcmClient: c,
	}, nil
}

type fcmImpl struct {
	*fcm.FcmClient
}

const errKey = "error"
const nrmsg = "NotRegistered"

// err表示錯誤內容
// errTokens表示非有效的token，需要做後續的處置
func (f *fcmImpl) Push(tokens []string, title, body string, data map[string]interface{}) (err error, errTokens []string) {
	f.SetNotificationPayload(&fcm.NotificationPayload{
		Title: title,
		Body:  body,
		Sound: "bingbong.aiff",
		Badge: "1",
	})
	data["title"] = title
	data["body"] = body
	f.AppendDevices(tokens)

	f.SetMsgData(data)
	r, err := f.Send()

	if err != nil {
		return err, nil
	}
	if r.Fail == 0 {
		return nil, nil
	}

	resultLen := len(r.Results)
	var msg string
	var ok bool
	for i := 0; i < resultLen; i++ {
		if msg, ok = r.Results[i][errKey]; ok && msg == nrmsg {
			errTokens = append(errTokens, tokens[i])
		}
	}
	return
}
