package log

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/94peter/sterna/util"
	"github.com/fluent/fluent-logger-golang/fluent"
)

const LogTargetFluent = "fluent"

type fluentLog struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Timezone string `yaml:"timezone"`
	service  string
}

func (rl *fluentLog) getClient() *fluent.Fluent {
	host := rl.Host
	port := rl.Port

	p, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	logger, err := fluent.New(fluent.Config{
		FluentHost: host,
		FluentPort: p,
	})
	if err != nil {
		return nil
	}
	return logger
}

var fluentTag = ""

func (rl *fluentLog) getTag() string {
	if fluentTag == "" {
		fluentTag = rl.service + "-log"
	}
	return fluentTag
}

func (rl *fluentLog) Write(p []byte) (n int, err error) {
	splitMsg := strings.SplitN(string(p), " ", 5)

	msg := map[string]interface{}{
		"service":  rl.service,
		"severity": splitMsg[0],
		"timeUnix": getTimeUnix(splitMsg[1], splitMsg[2], rl.Timezone),
		"filePath": splitMsg[3],
		"message":  strings.TrimSuffix(splitMsg[4], "\n"),
	}

	logger := rl.getClient()
	if logger == nil {
		return 0, nil
	}
	defer logger.Close()
	err = logger.Post(rl.getTag(), msg)
	if err != nil {
		fmt.Println(err)
	}
	return 0, nil
}

func getTimeUnix(d, t, z string) int64 {
	tt, err := time.Parse("2006/01/02 15:04:05.000000 -0700", util.StrAppend(d, " ", t, " ", z))
	if err != nil {
		return 0
	}
	return tt.Unix()
}
