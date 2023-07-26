package event

import (
	"fmt"
	"time"

	"github.com/94peter/sterna/log"
	mymqtt "github.com/94peter/sterna/mqtt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMqttEvent(service string, mqttConf *mymqtt.MqttConf, l log.Logger) Event {
	return &mqttEvent{
		service:  service,
		mqttConf: mqttConf,
		log:      l,
	}
}

type mqttEvent struct {
	service  string
	mqttConf *mymqtt.MqttConf
	log      log.Logger
}

func (p *mqttEvent) Fire(topic string, datas [][]byte) error {
	mqttserv, err := p.mqttConf.NewMqttServ(p.service, p.log)
	if err != nil {
		return err
	}
	defer mqttserv.Disconnect()
	for _, data := range datas {
		mqttserv.PublishByte([]string{topic}, data, false)
	}
	return nil
}

func (p *mqttEvent) Register(jobs ...EventJob) {
	mqttserv, err := p.mqttConf.NewMqttServ(p.service, p.log)
	if err != nil {
		p.log.Err("new mqttServ fail: " + err.Error())
	}
	defer mqttserv.Disconnect()
	subMap := mymqtt.MqttSubSerMap{}
	for _, j := range jobs {
		p.log.Info("mqtt job: " + j.GetTopic())
		subMap.Add(&mqttSubSer{
			topic:   j.GetTopic(),
			handler: j.GetHandler(),
			l:       p.log,
		})
	}
	waitDuration := time.Second * 30
	for {
		err = mqttserv.SubscribeMultiple(subMap)
		if err != nil {
			p.log.Err("subscribe fail: " + err.Error())
		} else {
			p.log.Info(fmt.Sprintf("try connect after %v", waitDuration))
		}
		time.Sleep(waitDuration)
	}
}

type mqttSubSer struct {
	topic   string
	handler func(data []byte) error
	l       log.Logger
}

func (m *mqttSubSer) GetTopic() string {
	return m.topic
}

func (m *mqttSubSer) GetQos() byte {
	return byte(1)
}

func (m *mqttSubSer) Handler(client mqtt.Client, msg mqtt.Message) {
	if err := m.handler(msg.Payload()); err != nil {
		m.l.Err("handler fail: " + err.Error())
	}
	msg.Ack()
}
