package mid

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/94peter/sterna"
	"github.com/94peter/sterna/util"
)

func NewServiceMid(di interface{}, servUri, env string) Middle {
	return &servMiddle{
		servUri: servUri,
		di:      di,
		env:     env,
	}
}

type servMiddle struct {
	servUri string
	env     string
	di      interface{}
}

func (lm *servMiddle) GetName() string {
	return "service"
}

var (
	serviceDiMap = make(map[string]interface{})
)

const (
	CtxServDiKey = util.CtxKey("ServiceDI")
)

func (am *servMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			service := r.Header.Get("X-Service")
			if service == "" {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			var mydi interface{}
			var ok bool
			if mydi, ok = serviceDiMap[service]; !ok {
				val := reflect.ValueOf(am.di)
				if val.Kind() == reflect.Ptr {
					val = reflect.Indirect(val)
				}
				newValue := reflect.New(val.Type()).Interface()
				err := sterna.InitConfByUri(fmt.Sprintf(am.servUri, service, am.env), newValue)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				serviceDiMap[service] = newValue
				mydi = newValue
			}
			r = util.SetCtxKeyVal(r, CtxServDiKey, mydi)
			f(w, r)
		}
	}
}
