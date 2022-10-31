package mid

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/94peter/sterna/auth"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
	"github.com/gorilla/mux"
)

type TokenParserResult interface {
	Host() string
	Perms() []string
	Account() string
	Name() string
	Sub() string
	Target() string
}

type AuthTokenParser interface {
	Parser(jwt auth.JwtDI, token string) (TokenParserResult, error)
}

func NewBearerAuthMid(tokenParser AuthTokenParser) AuthMidInter {
	return &bearAuthMiddle{
		parser:   tokenParser.Parser,
		authMap:  make(map[string]uint8),
		groupMap: make(map[string][]auth.UserPerm),
	}
}

func (lm *bearAuthMiddle) GetName() string {
	return "auth"
}

type bearAuthMiddle struct {
	parser   func(jwt auth.JwtDI, token string) (TokenParserResult, error)
	log      log.Logger
	authMap  map[string]uint8
	groupMap map[string][]auth.UserPerm
}

const (
	BearerAuthTokenKey = "Authorization"
)

func (am *bearAuthMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
	value := uint8(0)
	if isAuth {
		value = value | authValue
	}
	key := getPathKey(path, method)
	am.authMap[key] = uint8(value)
	am.groupMap[key] = group
}

func (am *bearAuthMiddle) IsAuth(path string, method string) bool {
	key := getPathKey(path, method)
	value, ok := am.authMap[key]
	if ok {
		return (value & authValue) > 0
	}
	return false
}

func (am *bearAuthMiddle) HasPerm(path, method string, perm []string) bool {
	key := fmt.Sprintf("%s:%s", path, method)
	groupAry, ok := am.groupMap[key]
	if !ok || groupAry == nil || len(groupAry) == 0 {
		return true
	}
	for _, g := range groupAry {
		if util.IsStrInList(string(g), perm...) {
			return true
		}
	}
	return false
}

func (am *bearAuthMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			path, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(err.Error()))
				return
			}
			if am.IsAuth(path, r.Method) {
				authToken := r.Header.Get(BearerAuthTokenKey)
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("miss token"))
					return
				}
				if !strings.HasPrefix(authToken, "Bearer ") {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("invalid token: missing Bearer"))
					return
				}
				servDi := util.GetCtxVal(r, CtxServDiKey)
				authToken = authToken[7:]
				result, err := am.parser(servDi.(auth.JwtDI), authToken)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("invalid token: " + err.Error()))
					return
				}
				if result.Host() != util.GetHost(r) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(fmt.Sprintf("host not match: [%s] is not [%s]", result.Host(), r.Host)))
					return
				}

				if hasPerm := am.HasPerm(path, r.Method, result.Perms()); !hasPerm {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("permission error"))
					return
				}
				r = util.SetCtxKeyVal(r, auth.CtxUserInfoKey, auth.NewReqUser(
					result.Host(),
					result.Sub(),
					result.Account(),
					result.Name(),
					result.Perms(),
				))
			}
			f(w, r)
		}
	}
}
