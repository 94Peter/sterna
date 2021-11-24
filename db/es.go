package db

import (
	"strings"
	"time"

	"github.com/94peter/sterna/util"
	elastic "github.com/olivere/elastic/v7"
)

const (
	dateFormat = "060102"
)

type EsDI interface {
	NewSearch() *elastic.Client
	GetIndexByTimeRange(startTime, endTime time.Time, typ string) string
	GetIndexByTime(d time.Time, _type string) string
}

type EsV7Conf struct {
	Hosts []string `yaml:"host"`
	Index string
}

func (s *EsV7Conf) NewSearch() *elastic.Client {
	myclient, err := elastic.NewClient(
		elastic.SetURL(s.Hosts...),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	return myclient
}

func (s *EsV7Conf) GetIndexByTimeRange(startTime, endTime time.Time, typ string) string {
	var searchIndex []string
	for startTime.Unix() <= endTime.Unix() {
		searchIndex = append(searchIndex, util.StrAppend(s.Index, "+", typ, "-", startTime.Format("060102")))
		startTime = startTime.AddDate(0, 0, 1)
		if startTime.YearDay() == endTime.YearDay() {
			searchIndex = append(searchIndex, util.StrAppend(s.Index, "+", typ, "-", startTime.Format("060102")))
		}
	}
	return strings.Join(searchIndex, ",")
}

func (s *EsV7Conf) GetIndexByTime(d time.Time, _type string) string {
	dk := d.Format(dateFormat)
	return util.StrAppend(s.Index, "+", _type, "-", dk)
}
