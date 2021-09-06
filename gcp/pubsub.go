package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sterna/log"
	"sterna/util"
	"sync"

	//"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

const (
	CtxPubSubKey = util.CtxKey("ctxPSKey")

	TopicQueue = "queue"
)

type AttrHandler interface {
	GetAttr() map[string]string
	Handler(msg *pubsub.Message)
}

type PubSubConf interface {
	NewPubSub(ctx context.Context, l log.Logger) PubSub
}

type SubHandler interface {
	GetSubID() string
	Handler(msg *pubsub.Message)
}

type PubSub interface {
	Publish(topicID string, msg []byte, attributes map[string]string) error
	Subcribe(sh SubHandler)
	Close()
}

func GetPubSubByReq(req *http.Request) PubSub {
	ctx := req.Context()
	cltInter := ctx.Value(CtxPubSubKey)
	if psclt, ok := cltInter.(PubSub); ok {
		return psclt
	}
	return nil
}

func (conf *GcpConf) NewPubSub(ctx context.Context, l log.Logger) PubSub {
	return &pubSubImpl{
		ctx: ctx,
		clt: conf.getPubSubClient(ctx),
		l:   l,
	}
}

func (conf *GcpConf) getPubSubClient(ctx context.Context) *pubsub.Client {
	var temp map[string]interface{}
	jsonFile, err := ioutil.ReadFile(conf.CredentialsFile)
	if err != nil {
		panic("ioutil.readfile error: " + err.Error())
	}
	err = json.Unmarshal(jsonFile, &temp)
	if err != nil {
		panic("unmarshal yaml err: " + err.Error())
	}

	client, err := pubsub.NewClient(ctx, temp["project_id"].(string), option.WithCredentialsFile(conf.CredentialsFile))
	if err != nil {
		panic("pubsub.NewClient error: " + err.Error())
	}
	return client
}

type pubSubImpl struct {
	clt *pubsub.Client
	ctx context.Context
	l   log.Logger
}

func (ps *pubSubImpl) Publish(topicID string, msg []byte, attributes map[string]string) error {
	t := ps.clt.Topic(topicID)
	result := t.Publish(ps.ctx, &pubsub.Message{
		Data:       msg,
		Attributes: attributes,
	})
	id, err := result.Get(ps.ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	ps.l.Info(fmt.Sprintf("Published message with custom attributes; msg ID: %v\n", id))
	return nil
}

// add a param: function, let go-routine to use it
func (ps *pubSubImpl) Subcribe(sh SubHandler) {
	var mu sync.Mutex
	sub := ps.clt.Subscription(sh.GetSubID())
	err := sub.Receive(ps.ctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()
		go sh.Handler(msg)
		msg.Ack()
	})
	if err != nil {
		ps.l.Err("Receive: " + err.Error())
	}
}

func (ps *pubSubImpl) Close() {
	if ps.clt != nil {
		ps.clt.Close()
	}
}
