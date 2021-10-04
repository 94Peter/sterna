package mgom

import (
	"github.com/94peter/sterna/dao"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MgoDS interface {
	Exec(exec func(d dao.DocInter) error) error
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

func (mm *findDsImpl) Exec(exec func(d dao.DocInter) error) error {
	return mm.FindAndExec(mm.d, mm.q, exec, mm.opts...)
}
