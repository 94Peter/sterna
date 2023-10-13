package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"
)

const (
	CtxLogKey = util.CtxKey("ctxLogKey")

	LogTargetOS = "os"

	infoPrefix  = "INFO (%s) "
	debugPrefix = "DEBUG (%s) "
	errorPrefix = "ERROR (%s) "
	warnPrefix  = "WARN (%s) "
	fatalPrefix = "FATAL (%s) "

	debugLevel = 1
	infoLevel  = 2
	warnLevel  = 3
	errorLevel = 4
	fatalLevel = 5
)

var (
	levelMap = map[string]int{
		"info":  infoLevel,
		"debug": debugLevel,
		"warn":  warnLevel,
		"error": errorLevel,
		"fatal": fatalLevel,
	}
)

func GetLogByReq(req *http.Request) Logger {
	return GetLogByCtx(req.Context())
}

func GetLogByCtx(ctx context.Context) Logger {
	cltInter := ctx.Value(CtxLogKey)

	if clt, ok := cltInter.(Logger); ok {
		return clt
	}
	return nil
}

func GetLogByGin(c *gin.Context) Logger {
	l, ok := c.Get(string(CtxLogKey))
	if !ok {
		return nil
	}
	return l.(Logger)
}

type Logger interface {
	Info(msg string)
	Debug(msg string)
	Warn(msg string)
	Err(msg string)
	Fatal(msg string)
}

type LoggerDI interface {
	NewLogger(key string) Logger
}

type LoggerConf struct {
	Level     string     `yaml:"-"`
	Target    string     `yaml:"-"`
	FluentLog *fluentLog `yaml:"fluent,omitempty"`
}

func (lc *LoggerConf) NewLogger(key string) Logger {
	if lc == nil {
		panic("log not set")
	}
	lc.Level = os.Getenv("LOG_LEVEL")
	if lc.Level == "" {
		lc.Level = "info"
	}
	lc.Target = os.Getenv("LOG_TARGET")
	if lc.Target == "" {
		lc.Target = LogTargetOS
	}
	level, ok := levelMap[lc.Level]
	if !ok {
		level = 0
	}
	myLevel := level

	var out io.Writer

	switch lc.Target {
	case LogTargetFluent:
		if lc.FluentLog == nil {
			panic("log fluent not set")
		}
		lc.FluentLog.service = key
		out = lc.FluentLog
	default:
		out = os.Stdout
	}

	return logImpl{
		logging: log.New(out, infoPrefix, log.Ldate|log.Lmicroseconds|log.Llongfile),
		key:     key,
		myLevel: myLevel,
	}
}

type logImpl struct {
	logging *log.Logger
	key     string
	myLevel int
}

func (l logImpl) Info(msg string) {
	if l.myLevel > infoLevel {
		return
	}
	l.logging.SetPrefix(fmt.Sprintf(infoPrefix, l.key))
	l.logging.Output(2, msg)
}

func (l logImpl) Debug(msg string) {
	if l.myLevel > debugLevel {
		return
	}
	l.logging.SetPrefix(fmt.Sprintf(debugPrefix, l.key))
	l.logging.Output(2, msg)
}

func (l logImpl) Warn(msg string) {
	if l.myLevel > warnLevel {
		return
	}
	l.logging.SetPrefix(fmt.Sprintf(warnPrefix, l.key))
	l.logging.Output(2, msg)
}

func (l logImpl) Err(msg string) {
	if l.myLevel > errorLevel {
		return
	}
	l.logging.SetPrefix(fmt.Sprintf(errorPrefix, l.key))
	l.logging.Output(2, msg)
}

func (l logImpl) Fatal(msg string) {
	if l.myLevel > fatalLevel {
		return
	}
	l.logging.SetPrefix(fmt.Sprintf(fatalPrefix, l.key))
	l.logging.Output(2, msg)
}
