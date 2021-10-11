package dao

import (
	"errors"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LogUser interface {
	GetName() string
	GetAccount() string
}

type DocInter interface {
	GetC() string
	GetDoc() interface{}
	GetID() interface{}
	SetCreator(u LogUser)
	AddRecord(u LogUser, msg string) []*Record
	GetIndexes() []mongo.IndexModel
}

type ListDoc interface {
	GetID() string
	GetC() string
	GetList(slice interface{}) []ListDoc
	BeforeSave() error
}

type CommonDoc struct {
	Records []*Record
}

func NewRecord(date time.Time, acc, name, msg string) *Record {
	return &Record{
		Datetime: date,
		Account:  acc,
		Name:     name,
		Summary:  msg,
	}
}

type Record struct {
	Datetime time.Time
	Summary  string
	Account  string
	Name     string
}

func (c *CommonDoc) AddRecord(u LogUser, msg string) []*Record {
	c.Records = append(c.Records, &Record{
		Datetime: time.Now(),
		Summary:  msg,
		Account:  u.GetAccount(),
		Name:     u.GetName(),
	})
	return c.Records
}

func (c *CommonDoc) SetCreator(lu LogUser) {
	if c == nil {
		return
	}
	c.Records = append(c.Records, &Record{
		Datetime: time.Now(),
		Summary:  "create",
		Account:  lu.GetAccount(),
		Name:     lu.GetName(),
	})
}

func (u *CommonDoc) GetC() string {
	panic("must override")
}

func Format(inter interface{}, f func(i interface{}) map[string]interface{}) (interface{}, int) {
	ik := reflect.TypeOf(inter).Kind()
	if ik == reflect.Ptr {
		return f(inter), 1
	}
	if ik != reflect.Slice {
		return nil, 0
	}
	v := reflect.ValueOf(inter)
	l := v.Len()
	ret := make([]map[string]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = f(v.Index(i).Interface())
	}
	return ret, l
}

func GetObjectID(id interface{}) (primitive.ObjectID, error) {
	var myID primitive.ObjectID
	switch dtype := reflect.TypeOf(id).String(); dtype {
	case "string":
		str := id.(string)
		return primitive.ObjectIDFromHex(str)
	case "primitive.ObjectID":
		myID = id.(primitive.ObjectID)
	default:
		return primitive.NilObjectID, errors.New("not support type: " + dtype)
	}
	return myID, nil
}
