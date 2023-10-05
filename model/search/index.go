package search

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

func AddDocument(ctx context.Context, clt *elasticsearch.Client, dao SearchDao) error {
	body, err := dao.Body()
	if err != nil {
		return err
	}
	indexReqA := esapi.IndexRequest{
		Index:      dao.Index(),
		DocumentID: dao.Id(),
		Body:       body,
	}
	resp, err := indexReqA.Do(ctx, clt)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New(getResponseBody(resp.Body))
	}
	return nil
}
func BulkAddDocumentWithIndex(ctx context.Context, clt *elasticsearch.Client, index string, dao []SearchDao) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         index,            // The default index name
		Client:        clt,              // The Elasticsearch client
		NumWorkers:    2,                // The number of worker goroutines
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		return err
	}
	for _, d := range dao {

		body, err := d.Body()
		if err != nil {
			return err
		}
		err = bi.Add(
			ctx,
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "create",
				// DocumentID is the (optional) document ID
				DocumentID: d.Id(),
				// Body is an `io.Reader` with the payload
				Body: body,
			},
		)
		if err != nil {
			return err
		}
	}
	if err = bi.Close(ctx); err != nil {
		return fmt.Errorf("close bulk index fail: " + err.Error())
	}
	return nil
}

func BulkAddDocument(ctx context.Context, clt *elasticsearch.Client, dao []SearchDao) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      dao[0].Index(), // The default index name
		Client:     clt,            // The Elasticsearch client
		NumWorkers: 2,              // The number of worker goroutines

		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		return err
	}
	for _, d := range dao {

		body, err := d.Body()
		if err != nil {
			return err
		}
		err = bi.Add(
			ctx,
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "create",
				// DocumentID is the (optional) document ID
				DocumentID: d.Id(),
				// Body is an `io.Reader` with the payload
				Body: body,
			},
		)
		if err != nil {
			return err
		}
	}
	if err = bi.Close(ctx); err != nil {
		return fmt.Errorf("close bulk index fail: " + err.Error())
	}
	return nil
}

func DeleteDocument(ctx context.Context, clt *elasticsearch.Client, dao SearchDao) error {
	deleteReq := esapi.DeleteRequest{
		Index:      dao.Index(),
		DocumentID: dao.Id(),
	}
	resp, err := deleteReq.Do(ctx, clt)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(getResponseBody(resp.Body))
	}
	return nil
}

func UpdateDucument(ctx context.Context, clt *elasticsearch.Client, dao SearchDao) error {
	body, err := dao.Body()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	bodyTpl := []byte(`{"doc": %s,"doc_as_upsert": true}`)
	newBody := bytes.Replace(
		bodyTpl, []byte("%s"), buf.Bytes(), 1)
	updateReq := esapi.UpdateRequest{
		Index:      dao.Index(),
		DocumentID: dao.Id(),
		Body:       bytes.NewReader(newBody),
	}
	resp, err := updateReq.Do(ctx, clt)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New(getResponseBody(resp.Body))
	}
	return err
}

func getResponseBody(rc io.ReadCloser) string {
	body, _ := ioutil.ReadAll(rc)
	rc.Close()
	return string(body)
}
