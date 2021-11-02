package db

import (
	"context"

	elastic "github.com/olivere/elastic/v7"
)

type EsDI interface {
	NewSearch(ctx context.Context) *elastic.Client
}

type EsV7Conf struct {
	Hosts []string `yaml:"host"`
	Index string
}

func (s *EsV7Conf) NewSearch(ctx context.Context) *elastic.Client {
	myclient, err := elastic.NewClient(
		elastic.SetURL(s.Hosts...),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	return myclient
}
