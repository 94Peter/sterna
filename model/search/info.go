package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
)

func Info(ctx context.Context, clt *elasticsearch.Client) (*InfoResult, error) {
	res, err := clt.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		return nil, errors.New(res.String())
	}
	var r map[string]interface{}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	return &InfoResult{
		ClientVersion: elasticsearch.Version,
		ServerVersion: r["version"].(map[string]interface{})["number"].(string),
	}, nil
}

type InfoResult struct {
	ClientVersion string
	ServerVersion string
}
