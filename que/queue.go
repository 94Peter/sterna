package queue

import (
	"fmt"
	"time"

	machinery "github.com/RichardKnop/machinery/v1"
	amqpConf "github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
)

type QueueDI interface {
	NewServ() (QueueServ, error)
}

type RabbitMQConf struct {
	*amqpConf.Config `yaml:",inline"`
}

func (conf *RabbitMQConf) NewServ() (QueueServ, error) {
	taskServer, err := machinery.NewServer(conf.Config)
	if err != nil {
		return nil, err
	}
	return &rabbitMQimpl{Server: taskServer}, nil
}

type QueueServ interface {
	RegisterTasks(workers ...Worker) error
	StartWorker() error
	SendTask(taskName string, args []tasks.Arg) (*tasks.TaskState, error)
}

type rabbitMQimpl struct {
	*machinery.Server
}

func (qs *rabbitMQimpl) fefresh() error {
	c := qs.Server.GetConfig()
	ms, err := machinery.NewServer(c)
	qs.Server = ms
	return err
}

func (qs *rabbitMQimpl) RegisterTasks(workers ...Worker) error {
	taskMap := make(map[string]interface{})
	for _, t := range workers {
		if t.Enable() {
			t.Register(taskMap)
		}
	}
	return qs.Server.RegisterTasks(taskMap)
}

func (qs *rabbitMQimpl) StartWorker() error {
	worker := qs.Server.NewWorker(fmt.Sprintf("worker-%d", time.Now().UnixNano()), 1)
	err := worker.Launch()
	if err != nil {
		worker.Quit()
		return err
	}
	return nil
}

func (qs *rabbitMQimpl) CountPendingTask(taskName string) (int, error) {
	// https://github.com/michaelklishin/rabbit-hole
	return 1, nil
}

func (qs *rabbitMQimpl) SendTask(taskName string, args []tasks.Arg) (*tasks.TaskState, error) {
	signature := &tasks.Signature{
		Name: taskName,
		Args: args,
	}
	asyncResult, err := qs.Server.SendTask(signature)
	if err != nil {
		qs.fefresh()
		return nil, err
	}
	return asyncResult.GetState(), nil
}

type Worker interface {
	Register(tasks map[string]interface{})
	Enable() bool
}
