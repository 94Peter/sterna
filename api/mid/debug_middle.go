package mid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"

	"github.com/gorilla/mux"
)

func NewDebugMid(name string) Middle {
	return &debugMiddle{
		name: name,
	}
}

func NewGinDebugMid(service string) GinMiddle {
	return &debugMiddle{
		name: service,
	}
}

type debugMiddle struct {
	name string
}

func (lm *debugMiddle) GetName() string {
	return lm.name
}

func (lm *debugMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log := log.GetLogByReq(r)
			if log == nil {
				f(w, r)
				return
			}
			log.Debug("-------Debug Request-------")
			path, _ := mux.CurrentRoute(r).GetPathTemplate()
			path = fmt.Sprintf("%s,%s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			log.Debug("path: " + path)
			header, _ := json.Marshal(r.Header)
			log.Debug("header: " + string(header))
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Err(fmt.Sprintf("Error reading body: %v", err))
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			log.Debug("body: " + string(b))
			start := time.Now()
			f(w, r)
			util.ListContextValue(r.Context(), false)
			delta := time.Now().Sub(start)
			if delta.Seconds() > 3 {
				log.Warn("too slow")
			}
			log.Debug("-------End Debug Request-------")
		}
	}
}

func (m *debugMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {

		fmt.Println("-------Request-------")
		path := c.FullPath()
		fmt.Println("full path: " + path)
		path = fmt.Sprintf("%s,%s?%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
		fmt.Println("path: " + path)
		header, _ := json.Marshal(c.Request.Header)
		fmt.Println("header: " + string(header))
		b, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("read body fail: ", err)
		}
		c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		fmt.Println("body: " + string(b))
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		start := time.Now()
		c.Next()
		delta := time.Now().Sub(start)
		if delta.Seconds() > 3 {
			fmt.Println("!!!! too slow !!!")
		}
		fmt.Println("-------End Request-------")
		fmt.Println("-------Response-------")
		fmt.Println(c.Writer.Status())
		header, _ = json.Marshal(c.Writer.Header())
		fmt.Println("header: " + string(header))
		fmt.Println("Response body: " + blw.body.String())
		fmt.Println("-------End Response-------")
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
