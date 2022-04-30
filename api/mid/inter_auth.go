package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/94peter/sterna/auth"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
	"github.com/gorilla/mux"
)

func NewInterAuthMid(url string) AuthMidInter {
	return &interAuthMiddle{
		url:      url,
		authMap:  make(map[string]uint8),
		groupMap: make(map[string][]auth.UserPerm),
	}
}

func (lm *interAuthMiddle) GetName() string {
	return "auth"
}

type interAuthMiddle struct {
	url      string
	log      log.Logger
	authMap  map[string]uint8
	groupMap map[string][]auth.UserPerm
}

func (am *interAuthMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
	value := uint8(0)
	if isAuth {
		value = value | authValue
	}
	key := getPathKey(path, method)
	am.authMap[key] = uint8(value)
	am.groupMap[key] = group
}

func (am *interAuthMiddle) IsAuth(path string, method string) bool {
	key := getPathKey(path, method)
	value, ok := am.authMap[key]
	if ok {
		return (value & authValue) > 0
	}
	return false
}

func (am *interAuthMiddle) HasPerm(path, method string, perm []string) bool {
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

func (am *interAuthMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
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
				// 打api取得token內容
				result, err := getParserToken(util.GetHost(r), am.url, authToken)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(err.Error()))
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

func getParserToken(host, url, token string) (TokenParserResult, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"X-Service":     []string{host},
		"Authorization": []string{token},
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(data))
	}

	pr := parseTokenResult{}
	err = json.NewDecoder(res.Body).Decode(&pr)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

type parseTokenResult map[string]interface{}

func (pr parseTokenResult) Account() string {
	return pr["account"].(string)
}

func (pr parseTokenResult) Host() string {
	return pr["host"].(string)
}

func (pr parseTokenResult) Name() string {
	return pr["name"].(string)
}

func (pr parseTokenResult) Perms() []string {
	return pr["perms"].([]string)
}

func (pr parseTokenResult) Sub() string {
	return pr["sub"].(string)
}
