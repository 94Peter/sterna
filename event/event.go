package event

import (
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/mqtt"
)

type EventJob interface {
	GetTopic() string
	GetHandler() EventHandler
}

type EventHandler func(data []byte) error

type Event interface {
	Fire(topic string, datas [][]byte) error
	Register(...EventJob)
}

type EventConf struct {
	Provider string
	Mqtt     *mqtt.MqttConf
}

func (ec *EventConf) NewEvent(serviceName string, l log.Logger) Event {
	if ec == nil {
		return nil
	}
	switch ec.Provider {
	case "mqtt":
		return NewMqttEvent(serviceName, ec.Mqtt, l)
	}
	return nil
}

type EventDI interface {
	NewEvent(serviceName string, l log.Logger) Event
}
