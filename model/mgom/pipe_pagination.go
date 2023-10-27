package mgom

import (
	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (mm *mgoModelImpl) GetPipePaginationSource(aggr MgoAggregate, q bson.M, sort bson.M) util.PaginationSource {
	return &mongoPipePaginationImpl{
		MgoDBModel: mm,
		a:          aggr,
		q:          q,
		sort:       sort,
	}
}

type mongoPipePaginationImpl struct {
	MgoDBModel
	a    MgoAggregate
	q    bson.M
	sort bson.M
}

func (mpi *mongoPipePaginationImpl) Count() (int64, error) {
	return mpi.CountDocuments(mpi.a, mpi.q)
}

func (mpi *mongoPipePaginationImpl) Data(limit, p int64, format func(i interface{}) map[string]interface{}) ([]map[string]interface{}, error) {
	result, err := mpi.PagePipeFind(mpi.a, mpi.q, mpi.sort, limit, p)
	if err != nil {
		return nil, err
	}
	formatResult, l := dao.Format(result, format)
	if l == 0 {
		return nil, nil
	}
	return formatResult.([]map[string]interface{}), nil
}
