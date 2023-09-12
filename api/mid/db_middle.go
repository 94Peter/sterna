package mid

import (
	"net/http"
	"runtime"

	"github.com/94peter/sterna"
	apiErr "github.com/94peter/sterna/api/err"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
)

type DBMidDI interface {
	log.LoggerDI
	db.MongoDI
	db.RedisDI
}

type DBMiddle string

func NewDBMid() Middle {
	return &dbMiddle{}
}

func NewGinDBMid(service string) GinMiddle {
	return &dbMiddle{
		service: service,
	}
}

type dbMiddle struct {
	service string
}

func (lm *dbMiddle) GetName() string {
	return "db"
}

func (lm *dbMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, lm.service, err)
}

func (am *dbMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			servDi := util.GetCtxVal(r, sterna.CtxServDiKey)
			if servDi == nil {
				apiErr.OutputErr(w, apiErr.New(http.StatusInternalServerError, "can not get di"))
				return
			}

			if dbdi, ok := servDi.(DBMidDI); ok {
				uuid := uuid.New().String()
				l := dbdi.NewLogger(uuid)

				dbclt, err := dbdi.NewMongoDBClient(r.Context(), "")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				defer dbclt.Close()
				r = util.SetCtxKeyVal(r, db.CtxMongoKey, dbclt)
				r = util.SetCtxKeyVal(r, log.CtxLogKey, l)
				f(w, r)
				runtime.GC()
			} else {
				apiErr.OutputErr(w, apiErr.New(http.StatusInternalServerError, "invalid di"))
				return
			}
		}
	}
}

func (m *dbMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {

		servDi, ok := c.Get(string(sterna.CtxServDiKey))
		if !ok || servDi == nil {
			c.Next()
			return
		}

		if dbdi, ok := servDi.(DBMidDI); ok {
			uuid := uuid.New().String()
			l := dbdi.NewLogger(uuid)

			dbclt, err := dbdi.NewMongoDBClient(c.Request.Context(), "")
			if err != nil {
				m.outputErr(c, apiErr.New(http.StatusInternalServerError, err.Error()))
				c.Abort()
				return
			}
			defer dbclt.Close()

			c.Set(string(db.CtxMongoKey), dbclt)
			c.Set(string(log.CtxLogKey), l)

			c.Next()
			runtime.GC()
		} else {
			m.outputErr(c, apiErr.New(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}
