package search

import (
	"context"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func IsIndexExist(ctx context.Context, clt *elasticsearch.Client, index string) (bool, error) {
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{index},
	}
	resp, err := existsReq.Do(ctx, clt)
	if err != nil {
		return false, err
	}
	return resp.StatusCode != http.StatusNotFound, nil
}
