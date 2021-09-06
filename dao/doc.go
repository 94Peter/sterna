package dao

import (
	"errors"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	AddRecord(u LogUser, msg string) []*record
	//GetUpdateField() bson.M

	//GetSaveTxnOp(lu LogUser) txn.Op
	// GetUpdateTxnOp(data bson.D) txn.Op
	// GetDelTxnOp() txn.Op

	GetIndexes() []mongo.IndexModel
}

type ListDoc interface {
	GetID() string
	GetC() string
	GetList(slice interface{}) []ListDoc
	BeforeSave() error
}

type CommonDoc struct {
	Records []*record
}

type record struct {
	Datetime time.Time
	Summary  string
	Account  string
	Name     string
}

func (c *CommonDoc) AddRecord(u LogUser, msg string) []*record {
	c.Records = append(c.Records, &record{
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
	c.Records = append(c.Records, &record{
		Datetime: time.Now(),
		Summary:  "create",
		Account:  lu.GetAccount(),
		Name:     lu.GetName(),
	})
}

// func (u *CommonDoc) GetMongoIndexes() []mgo.Index {
// 	return nil
// }

// func (u *CommonDoc) getSaveTxnOp(doc DocInter, lu LogUser) txn.Op {
// 	u.SetCreator(lu)
// 	doc.SetCompany(lu.GetCompany())
// 	d := doc.GetDoc()
// 	return txn.Op{
// 		C:      doc.GetC(),
// 		Id:     doc.GetID(),
// 		Assert: txn.DocMissing,
// 		Insert: d,
// 	}
// }

// func (u *CommonDoc) getUpdateTxnOp(doc DocInter, data bson.D) txn.Op {
// 	return txn.Op{
// 		C:      doc.GetC(),
// 		Id:     doc.GetID(),
// 		Assert: txn.DocExists,
// 		Update: bson.D{{Name: "$set", Value: data}},
// 	}
// }

// func (u *CommonDoc) getDelTxnOp(doc DocInter) txn.Op {
// 	return txn.Op{
// 		C:      doc.GetC(),
// 		Id:     doc.GetID(),
// 		Assert: txn.DocExists,
// 		Remove: true,
// 	}
// }

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

type PipelineQryInter interface {
	GetQ() bson.M
	GetPipeline() []bson.M
}

type CommonPipeline struct {
	q        bson.M
	pipeline []bson.M
}

func (u *CommonPipeline) GetQ() bson.M {
	return u.q
}

func (u *CommonPipeline) GetPipeline() []bson.M {
	return u.pipeline
}
