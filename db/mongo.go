package db

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/94peter/sterna/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	CtxMongoKey = util.CtxKey("ctxMongoKey")
	HeaderDBKey = "raccMongoDB"
)

func GetCtxMgoDBClient(req *http.Request) MongoDBClient {
	ctx := req.Context()
	cltInter := ctx.Value(CtxMongoKey)
	if dbclt, ok := cltInter.(MongoDBClient); ok {
		return dbclt
	}
	return nil
}

type MongoDI interface {
	NewMongoDBClient(ctx context.Context, userDB string) (MongoDBClient, error)
}

type MongoConf struct {
	Uri       string `yaml:"uri"`
	User      string `yaml:"user"`
	Pass      string `yaml:"pass"`
	DefaultDB string `yaml:"defaul"`

	lock     *sync.Mutex
	connPool map[string]*mongo.Client
}

func (mc *MongoConf) NewMongoDBClient(ctx context.Context, userDB string) (MongoDBClient, error) {
	if mc.Uri == "" {
		panic("mongo uri not set")
	}
	if mc.DefaultDB == "" {
		panic("mongo default db not set")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mc.Uri))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		cancel()
		return nil, err
	}

	return &mgoClientImpl{
		clt:       client,
		ctx:       ctx,
		cancel:    cancel,
		dbPool:    make(map[string]*mongo.Database),
		defaultDB: mc.DefaultDB,
		userDB:    userDB,
	}, nil

}

type mgoClientImpl struct {
	clt     *mongo.Client
	ctx     context.Context
	cancel  context.CancelFunc
	dbPool  map[string]*mongo.Database
	session mongo.Session

	defaultDB string
	userDB    string
}

func (m *mgoClientImpl) WithSession(f func(sc mongo.SessionContext) error) error {
	if m.session != nil {
		return nil
	}
	session, err := m.clt.StartSession()
	if err != nil {
		m.cancel()
		return err
	}
	if err := session.StartTransaction(); err != nil {
		return err
	}
	m.session = session
	return mongo.WithSession(m.ctx, m.session, f)
}

func (m *mgoClientImpl) GetDBList() ([]string, error) {
	return m.clt.ListDatabaseNames(m.ctx, bson.M{})
}

func (m *mgoClientImpl) getDB(db string) *mongo.Database {
	if db == "" {
		db = m.defaultDB
	}
	if dbclt, ok := m.dbPool[db]; ok {
		return dbclt
	}
	dbclt := m.clt.Database(db)
	m.dbPool[db] = dbclt
	return dbclt
}

func (m *mgoClientImpl) Close() {
	if m == nil {
		return
	}
	if m.session != nil {
		m.session.EndSession(m.ctx)
	}
	if m.clt != nil {
		err := m.clt.Disconnect(m.ctx)
		if err != nil {
			fmt.Println("disconnect error: " + err.Error())
		}
	}
	m.cancel()
}

func (m *mgoClientImpl) Ping() error {
	return m.clt.Ping(m.ctx, readpref.Primary())
}

func (m *mgoClientImpl) GetCoreDB() *mongo.Database {
	return m.getDB(m.defaultDB)
}

func (m *mgoClientImpl) GetUserDB() *mongo.Database {
	if m.userDB == "" {
		return nil
	}
	return m.getDB(m.userDB)
}

func (m *mgoClientImpl) AbortTransaction(sc mongo.SessionContext) error {
	return m.session.AbortTransaction(sc)
}
func (m *mgoClientImpl) CommitTransaction(sc mongo.SessionContext) error {
	return m.session.CommitTransaction(sc)
}

const (
	CoreDB = "core"
	UserDB = "user"
)

type MongoDBClient interface {
	GetCoreDB() *mongo.Database
	GetUserDB() *mongo.Database
	WithSession(f func(sc mongo.SessionContext) error) error
	AbortTransaction(sc mongo.SessionContext) error
	CommitTransaction(sc mongo.SessionContext) error
	Close()
}
