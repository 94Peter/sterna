package mgom

import (
	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mm *mgoModelImpl) GetPaginationSource(d dao.DocInter, q bson.M, opts ...*options.FindOptions) util.PaginationSource {
	return &mongoPaginationImpl{
		MgoDBModel: mm,
		d:          d,
		q:          q,
		findOpts:   opts,
	}
}

type mongoPaginationImpl struct {
	MgoDBModel
	d        dao.DocInter
	q        bson.M
	findOpts []*options.FindOptions
}

func (mpi *mongoPaginationImpl) Count() (int64, error) {
	return mpi.CountDocuments(mpi.d, mpi.q)
}

func (mpi *mongoPaginationImpl) Data(limit, p int64, format func(i interface{}) map[string]interface{}) ([]map[string]interface{}, error) {
	result, err := mpi.PageFind(mpi.d, mpi.q, limit, p, mpi.findOpts...)
	if err != nil {
		return nil, err
	}
	formatResult, l := dao.Format(result, format)
	if l == 0 {
		return nil, nil
	}
	return formatResult.([]map[string]interface{}), nil
}
