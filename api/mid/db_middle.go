package mid

import (
	"net/http"
	"runtime"

	"sterna/api/input"
	"sterna/db"
	"sterna/log"
	"sterna/util"

	"github.com/google/uuid"
)

type DBMidDI interface {
	log.LoggerDI
	db.MongoDI
}

type DBMiddle string

func NewDBMid(di DBMidDI, name string) Middle {
	return &dbMiddle{
		name: name,
		di:   di,
	}
}

type dbMiddle struct {
	name string
	di   DBMidDI
}

func (lm *dbMiddle) GetName() string {
	return lm.name
}

func (am *dbMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			uuid := uuid.New().String()
			l := am.di.NewLogger(uuid)
			userInfo := input.GetUserInfo(r)
			userDB := ""
			if userInfo != nil && userInfo.GetDB() != "" {
				userDB = "ws:" + userInfo.GetDB()
			}
			dbclt, err := am.di.NewMongoDBClient(r.Context(), userDB)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			r = util.SetCtxKeyVal(r, db.CtxMongoKey, dbclt)
			r = util.SetCtxKeyVal(r, log.CtxLogKey, l)
			f(w, r)
			dbclt.Close()
			runtime.GC()
		}
	}
}
