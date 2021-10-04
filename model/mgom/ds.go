package mgom

import (
	"github.com/94peter/sterna/dao"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MgoDS interface {
	Exec(exec func(i interface{}) error) error
}

func (mm *mgoModelImpl) NewFindMgoDS(d dao.DocInter, q bson.M, opts ...*options.FindOptions) MgoDS {
	return &findDsImpl{
		MgoDBModel: mm,
		d:          d,
		q:          q,
		opts:       opts,
	}
}

type findDsImpl struct {
	MgoDBModel
	d    dao.DocInter
	q    bson.M
	opts []*options.FindOptions
}

func (mm *findDsImpl) Exec(exec func(i interface{}) error) error {
	return mm.FindAndExec(mm.d, mm.q, exec, mm.opts...)
}

func (mm *mgoModelImpl) NewPipeFindMgoDS(d MgoAggregate, q bson.M, opts ...*options.AggregateOptions) MgoDS {
	return &pipeFindDsImpl{
		MgoDBModel: mm,
		d:          d,
		q:          q,
		opts:       opts,
	}
}

type pipeFindDsImpl struct {
	MgoDBModel
	d    MgoAggregate
	q    bson.M
	opts []*options.AggregateOptions
}

func (mm *pipeFindDsImpl) Exec(exec func(i interface{}) error) error {
	return mm.PipeFindAndExec(mm.d, mm.q, exec, mm.opts...)
}
