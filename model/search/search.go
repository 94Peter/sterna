package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
)

type Q map[string]any

type SearchDao interface {
	Id() string
	Index() string
	Body() (*bytes.Reader, error)
	GetMapping() string
}

type SearchResult struct {
	Status string
	Total  int
	Took   int
	Hits   []*hit
}

func (sr *SearchResult) AddHit(id string, s map[string]any) {
	sr.Hits = append(sr.Hits, &hit{ID: id, Source: s})
}

type hit struct {
	ID     string
	Source map[string]any
}

func Search(ctx context.Context, clt *elasticsearch.Client, dao SearchDao, q Q) (*SearchResult, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		return nil, err
	}
	res, err := clt.Search(
		clt.Search.WithContext(ctx),
		clt.Search.WithIndex(dao.Index()),
		clt.Search.WithBody(&buf),
		clt.Search.WithTrackTotalHits(true),
		clt.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		} else {
			// Print the response status and error information.
			return nil, fmt.Errorf(
				"[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing the response body: %s", err)
	}

	result := SearchResult{
		Status: res.Status(),
		Total:  int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		Took:   int(r["took"].(float64)),
	}

	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		result.AddHit(hit.(map[string]interface{})["_id"].(string), hit.(map[string]interface{})["_source"].(map[string]interface{}))
	}
	return &result, nil
}
