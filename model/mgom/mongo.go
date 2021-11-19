package mgom

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
)

type MgoAggregate interface {
	GetPipeline(q bson.M) mongo.Pipeline
	GetC() string
}

type MgoDBModel interface {
	SetDB(db *mongo.Database)
	BatchSave(doclist []dao.DocInter, u dao.LogUser) (inserted []interface{}, failed []dao.DocInter, err error)
	Save(d dao.DocInter, u dao.LogUser) (interface{}, error)
	RemoveAll(d dao.DocInter, q primitive.M, u dao.LogUser) (int64, error)
	RemoveByID(d dao.DocInter, u dao.LogUser) (int64, error)
	UpdateOne(d dao.DocInter, fields bson.D, u dao.LogUser) (int64, error)
	UpdateAll(d dao.DocInter, q bson.M, fields bson.D, u dao.LogUser) (int64, error)
	UnsetFields(d dao.DocInter, q bson.M, fields []string, u dao.LogUser) (int64, error)
	Upsert(d dao.DocInter, u dao.LogUser) (interface{}, error)
	FindByID(d dao.DocInter) error
	FindOne(d dao.DocInter, q bson.M, option ...*options.FindOneOptions) error
	Find(d dao.DocInter, q bson.M, option ...*options.FindOptions) (interface{}, error)
	FindAndExec(
		d dao.DocInter, q bson.M,
		exec func(i interface{}) error,
		opts ...*options.FindOptions,
	) error
	PipeFindOne(aggr MgoAggregate, filter bson.M) error
	PipeFind(aggr MgoAggregate, filter bson.M, opts ...*options.AggregateOptions) (interface{}, error)
	PipeFindAndExec(
		aggr MgoAggregate, q bson.M,
		exec func(i interface{}) error,
		opts ...*options.AggregateOptions,
	) error
	PagePipeFind(aggr MgoAggregate, filter bson.M, limit, page int64) (interface{}, error)
	PageFind(d dao.DocInter, q bson.M, limit, page int64) (interface{}, error)
	CountDocuments(d dao.DocInter, q bson.M) (int64, error)
	GetPaginationSource(d dao.DocInter, q bson.M) util.PaginationSource
	CreateCollection(dlist ...dao.DocInter) error

	//Reference to customer code, use for aggregate pagination
	AggrCountDocuments(aggr MgoAggregate, q bson.M) (int64, error)
	GetPipePaginationSource(aggr MgoAggregate, q bson.M) util.PaginationSource

	NewFindMgoDS(d dao.DocInter, q bson.M, opts ...*options.FindOptions) MgoDS
	NewPipeFindMgoDS(d MgoAggregate, q bson.M, opts ...*options.AggregateOptions) MgoDS
}

func NewMgoModel(ctx context.Context, db *mongo.Database, log log.Logger) MgoDBModel {
	return &mgoModelImpl{
		db:      db,
		ctx:     ctx,
		selfCtx: context.Background(),
		log:     log,
	}
}

func NewMgoModelByReq(req *http.Request, source string) MgoDBModel {
	mgodbclt := db.GetCtxMgoDBClient(req)
	if mgodbclt == nil {
		panic("database not set in req")
	}
	log := log.GetCtxLog(req)
	if log == nil {
		panic("log not set in req")
	}
	if source == db.CoreDB {
		return &mgoModelImpl{
			db:      mgodbclt.GetCoreDB(),
			ctx:     req.Context(),
			selfCtx: context.Background(),
			log:     log,
		}
	}
	udb := mgodbclt.GetUserDB()
	if udb == nil {
		panic("user not set")
	}
	return &mgoModelImpl{
		db:      udb,
		ctx:     req.Context(),
		selfCtx: context.Background(),
		log:     log,
	}
}

func GetObjectID(id interface{}) (primitive.ObjectID, error) {
	switch dtype := reflect.TypeOf(id).String(); dtype {
	case "string":
		str := id.(string)
		return primitive.ObjectIDFromHex(str)
	case "primitive.ObjectID":
		return id.(primitive.ObjectID), nil
	default:
		return primitive.NilObjectID, errors.New("not support type: " + dtype)
	}
}

type mgoModelImpl struct {
	db  *mongo.Database
	log log.Logger
	ctx context.Context

	selfCtx       context.Context
	indexExistMap map[string]bool
}

func (mm *mgoModelImpl) SetDB(db *mongo.Database) {
	mm.db = db
}

func (mm *mgoModelImpl) FindAndExec(
	d dao.DocInter, q bson.M,
	exec func(i interface{}) error,
	opts ...*options.FindOptions,
) error {
	var err error
	collection := mm.db.Collection(d.GetC())
	sortCursor, err := collection.Find(mm.ctx, q, opts...)
	if err != nil {
		return nil
	}
	val := reflect.ValueOf(d)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	var newValue reflect.Value
	var newDoc dao.DocInter
	for sortCursor.Next(mm.ctx) {
		newValue = reflect.New(val.Type())
		newDoc = newValue.Interface().(dao.DocInter)
		err = sortCursor.Decode(newDoc)
		if err != nil {
			return err
		}
		err = exec(newDoc)
		if err != nil {
			return err
		}
	}
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		if newValue.IsZero() {
			continue
		}
		f.Set(newValue.Elem().Field(i))
	}
	return err
}

func (mm *mgoModelImpl) CountDocuments(d dao.DocInter, q bson.M) (int64, error) {
	opts := options.Count().SetMaxTime(2 * time.Second)
	return mm.db.Collection(d.GetC()).CountDocuments(mm.ctx, q, opts)
}

func (mm *mgoModelImpl) isCollectExisted(d dao.DocInter) bool {
	names, err := mm.db.ListCollectionNames(mm.selfCtx, bson.D{{Key: "name", Value: d.GetC()}})
	if ce, ok := err.(mongo.CommandError); ok {
		if ce.Name == "OperationNotSupportedInTransaction" {
			return true
		}
		return false
	}

	return util.IsStrInList(d.GetC(), names...)
}

func (mm *mgoModelImpl) CreateCollection(dlist ...dao.DocInter) (err error) {
	var indexStr []string
	for _, d := range dlist {
		mm.log.Info("check collection " + d.GetC())
		if !mm.isCollectExisted(d) {
			if len(d.GetIndexes()) > 0 {
				indexStr, err = mm.db.Collection(d.GetC()).Indexes().CreateMany(mm.selfCtx, d.GetIndexes())
				mm.log.Info(fmt.Sprintln("created index: ", indexStr))
			} else {
				err = mm.db.CreateCollection(mm.selfCtx, d.GetC())
			}
			if err != nil {
				mm.log.Warn(fmt.Sprintf("created collection [%s] fail: %s", d.GetC(), err.Error()))
			} else {
				mm.log.Info("collection created: " + d.GetC())
			}
		}
	}
	return
}

func (mm *mgoModelImpl) BatchSave(doclist []dao.DocInter, u dao.LogUser) (inserted []interface{}, failed []dao.DocInter, err error) {
	if len(doclist) == 0 {
		inserted = nil
		return
	}
	collection := mm.db.Collection(doclist[0].GetC())
	ordered := false
	var batch []interface{}
	for _, d := range doclist {
		if u != nil {
			d.SetCreator(u)
		}
		batch = append(batch, d)
	}
	var result *mongo.InsertManyResult
	result, err = collection.InsertMany(mm.ctx, batch, &options.InsertManyOptions{Ordered: &ordered})
	if result != nil {
		inserted = result.InsertedIDs
	}

	if excep, ok := err.(mongo.BulkWriteException); ok {
		for _, e := range excep.WriteErrors {
			failed = append(failed, doclist[e.Index])
		}
	}
	return
}

func (mm *mgoModelImpl) Save(d dao.DocInter, u dao.LogUser) (interface{}, error) {
	err := mm.CreateCollection(d)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if u != nil {
		d.SetCreator(u)
	}
	collection := mm.db.Collection(d.GetC())

	result, err := collection.InsertOne(mm.ctx, d.GetDoc())
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID, err

}

func (mm *mgoModelImpl) RemoveAll(d dao.DocInter, q primitive.M, u dao.LogUser) (int64, error) {
	collection := mm.db.Collection(d.GetC())
	result, err := collection.DeleteMany(mm.ctx, q)
	return result.DeletedCount, err
}

func (mm *mgoModelImpl) RemoveByID(d dao.DocInter, u dao.LogUser) (int64, error) {
	collection := mm.db.Collection(d.GetC())
	result, err := collection.DeleteOne(mm.ctx, bson.M{"_id": d.GetID()})
	return result.DeletedCount, err
}

func (mm *mgoModelImpl) UpdateOne(d dao.DocInter, fields bson.D, u dao.LogUser) (int64, error) {
	if u != nil {
		fields = append(fields, primitive.E{Key: "records", Value: d.AddRecord(u, "updated")})
	}
	collection := mm.db.Collection(d.GetC())
	result, err := collection.UpdateOne(mm.ctx, bson.M{"_id": d.GetID()},
		bson.D{
			{Key: "$set", Value: fields},
		},
	)
	if result != nil {
		return result.ModifiedCount, err
	}
	return 0, err
}

func (mm *mgoModelImpl) UpdateAll(d dao.DocInter, q bson.M, fields bson.D, u dao.LogUser) (int64, error) {
	if u != nil {
		fields = append(fields, primitive.E{Key: "records", Value: d.AddRecord(u, "updated")})
	}
	collection := mm.db.Collection(d.GetC())
	result, err := collection.UpdateMany(mm.ctx, q,
		bson.D{
			{Key: "$set", Value: fields},
		},
	)
	if result != nil {
		return result.ModifiedCount, err
	}
	return 0, err
}

func (mm *mgoModelImpl) UnsetFields(d dao.DocInter, q bson.M, fields []string, u dao.LogUser) (int64, error) {
	collection := mm.db.Collection(d.GetC())
	m := primitive.M{}
	for _, k := range fields {
		m[k] = ""
	}
	result, err := collection.UpdateMany(mm.ctx, q,
		bson.D{
			{Key: "$unset", Value: m},
		},
	)
	if result != nil {
		return result.ModifiedCount, err
	}
	return 0, err
}

func (mm *mgoModelImpl) Upsert(d dao.DocInter, u dao.LogUser) (interface{}, error) {
	err := mm.CreateCollection(d)
	if err != nil {
		return primitive.NilObjectID, err
	}

	collection := mm.db.Collection(d.GetC())
	_, err = collection.UpdateOne(mm.ctx, bson.M{"_id": d.GetID()}, bson.M{"$set": d.GetDoc()}, options.Update().SetUpsert(true))

	if err != nil {
		return primitive.NilObjectID, err
	}
	return d.GetID(), nil
}

func (mm *mgoModelImpl) FindByID(d dao.DocInter) error {
	return mm.FindOne(d, bson.M{"_id": d.GetID()})
}

func (mm *mgoModelImpl) FindOne(d dao.DocInter, q bson.M, option ...*options.FindOneOptions) error {
	if mm.db == nil {
		return errors.New("db is nil")
	}
	if d == nil {
		return errors.New("doc is nil")
	}
	collection := mm.db.Collection(d.GetC())
	return collection.FindOne(mm.ctx, q, option...).Decode(d)
}

func (mm *mgoModelImpl) Find(d dao.DocInter, q bson.M, option ...*options.FindOptions) (interface{}, error) {
	myType := reflect.TypeOf(d)
	slice := reflect.MakeSlice(reflect.SliceOf(myType), 0, 0).Interface()
	collection := mm.db.Collection(d.GetC())
	sortCursor, err := collection.Find(mm.ctx, q, option...)
	if err != nil {
		return nil, err
	}
	err = sortCursor.All(mm.ctx, &slice)
	if err != nil {
		return nil, err
	}
	return slice, err
}

func (mm *mgoModelImpl) PipeFind(aggr MgoAggregate, filter bson.M, opts ...*options.AggregateOptions) (interface{}, error) {
	myType := reflect.TypeOf(aggr)
	slice := reflect.MakeSlice(reflect.SliceOf(myType), 0, 0).Interface()
	collection := mm.db.Collection(aggr.GetC())
	sortCursor, err := collection.Aggregate(mm.ctx, aggr.GetPipeline(filter), opts...)
	if err != nil {
		return nil, err
	}
	err = sortCursor.All(mm.ctx, &slice)
	if err != nil {
		return nil, err
	}
	return slice, err
}

func (mm *mgoModelImpl) PipeFindAndExec(aggr MgoAggregate, filter bson.M, exec func(i interface{}) error, opts ...*options.AggregateOptions) error {
	if aggr == nil {
		return errors.New("aggr is nil")
	}
	collection := mm.db.Collection(aggr.GetC())
	sortCursor, err := collection.Aggregate(mm.ctx, aggr.GetPipeline(filter), opts...)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(aggr)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	var newValue reflect.Value
	var newDoc dao.DocInter
	for sortCursor.Next(mm.ctx) {
		newValue = reflect.New(val.Type())
		newDoc = newValue.Interface().(dao.DocInter)
		err = sortCursor.Decode(newDoc)
		fmt.Println(newDoc)
		if err != nil {
			return err
		}
		err = exec(newDoc)
		if err != nil {
			return err
		}
	}
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		if newValue.IsNil() || !newValue.IsValid() || newValue.IsZero() {
			continue
		}
		f.Set(newValue.Elem().Field(i))
	}
	return err
}

func (mm *mgoModelImpl) PipeFindOne(aggr MgoAggregate, filter bson.M) error {
	collection := mm.db.Collection(aggr.GetC())
	sortCursor, err := collection.Aggregate(mm.ctx, aggr.GetPipeline(filter))
	if err != nil {
		return err
	}
	if sortCursor.Next(mm.ctx) {
		err = sortCursor.Decode(aggr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mm *mgoModelImpl) PageFind(d dao.DocInter, filter bson.M, limit, page int64) (interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	skip := limit * (page - 1)
	findopt := options.Find().SetSkip(skip).SetLimit(limit)
	myType := reflect.TypeOf(d)
	slice := reflect.MakeSlice(reflect.SliceOf(myType), 0, 0).Interface()
	collection := mm.db.Collection(d.GetC())
	sortCursor, err := collection.Find(mm.ctx, filter, findopt)
	if err != nil {
		return nil, err
	}

	err = sortCursor.All(mm.ctx, &slice)
	return slice, err
}

func (mm *mgoModelImpl) PagePipeFind(aggr MgoAggregate, filter bson.M, limit, page int64) (interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	skip := limit * (page - 1)
	myType := reflect.TypeOf(aggr)
	slice := reflect.MakeSlice(reflect.SliceOf(myType), 0, 0).Interface()

	collection := mm.db.Collection(aggr.GetC())
	pl := append(aggr.GetPipeline(filter), bson.D{{Key: "$skip", Value: skip}}, bson.D{{Key: "$limit", Value: limit}})
	sortCursor, err := collection.Aggregate(mm.ctx, pl)
	if err != nil {
		return nil, err
	}
	err = sortCursor.All(mm.ctx, &slice)
	if err != nil {
		return nil, err
	}
	return slice, err
}

func (mm *mgoModelImpl) GetPaginationSource(d dao.DocInter, q bson.M) util.PaginationSource {
	return &mongoPaginationImpl{
		MgoDBModel: mm,
		d:          d,
		q:          q,
	}
}

type mongoPaginationImpl struct {
	MgoDBModel
	d dao.DocInter
	q bson.M
}

func (mpi *mongoPaginationImpl) Count() (int64, error) {
	return mpi.CountDocuments(mpi.d, mpi.q)
}

func (mpi *mongoPaginationImpl) Data(limit, p int64, format func(i interface{}) map[string]interface{}) ([]map[string]interface{}, error) {
	result, err := mpi.PageFind(mpi.d, mpi.q, limit, p)
	if err != nil {
		return nil, err
	}
	formatResult, l := dao.Format(result, format)
	if l == 0 {
		return nil, nil
	}
	return formatResult.([]map[string]interface{}), nil
}

// ----- New added code -----

func (mm *mgoModelImpl) GetPipePaginationSource(aggr MgoAggregate, q bson.M) util.PaginationSource {
	return &mongoPipePaginationImpl{
		MgoDBModel: mm,
		a:          aggr,
		q:          q,
	}
}

func (mm *mgoModelImpl) AggrCountDocuments(aggr MgoAggregate, q bson.M) (int64, error) {
	opts := options.Count().SetMaxTime(2 * time.Second)
	return mm.db.Collection(aggr.GetC()).CountDocuments(mm.ctx, q, opts)
}

type mongoPipePaginationImpl struct {
	MgoDBModel
	a MgoAggregate
	q bson.M
}

func (mpi *mongoPipePaginationImpl) Count() (int64, error) {
	return mpi.AggrCountDocuments(mpi.a, mpi.q)
}

func (mpi *mongoPipePaginationImpl) Data(limit, p int64, format func(i interface{}) map[string]interface{}) ([]map[string]interface{}, error) {
	result, err := mpi.PagePipeFind(mpi.a, mpi.q, limit, p)
	if err != nil {
		return nil, err
	}
	formatResult, l := dao.Format(result, format)
	if l == 0 {
		return nil, nil
	}
	return formatResult.([]map[string]interface{}), nil
}
