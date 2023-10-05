package search

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func CreateIndexByDao(ctx context.Context, clt *elasticsearch.Client, dao SearchDao) error {
	indexReq := esapi.IndicesCreateRequest{
		Index: dao.Index(),
		Body:  strings.NewReader(dao.GetMapping()),
	}
	_, err := indexReq.Do(ctx, clt)
	if err != nil {
		return err
	}
	return nil
}

func CreateIndex(ctx context.Context, clt *elasticsearch.Client, index string, mapping string) error {
	indexReq := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(mapping),
	}
	_, err := indexReq.Do(ctx, clt)
	if err != nil {
		return err
	}
	return nil
}
