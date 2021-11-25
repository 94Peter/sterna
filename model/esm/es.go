package esm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/log"

	elastic "github.com/olivere/elastic/v7"
)

type EsSearch interface {
	BulkByDate(sd dao.SearchDaoList, d time.Time) error
	PutByDate(sd dao.ESDao, t time.Time) error
}

type searchModel struct {
	ctx      context.Context
	esDI     db.EsDI
	search   *elastic.Client
	indexMap map[string]bool
	log      log.Logger

	*sync.Mutex
}

func NewSearchModel(ctx context.Context, c db.EsDI, l log.Logger) EsSearch {
	return &searchModel{
		ctx:    ctx,
		esDI:   c,
		search: c.NewSearch(),
		log:    l,
	}
}

func (sm *searchModel) BulkByDate(sd dao.SearchDaoList, d time.Time) error {
	l := len(sd)
	if l == 0 {
		return errors.New("there is no searchDao")
	}
	ty := sd[0].GetType()
	index := sm.esDI.GetIndexByTime(d, ty)
	if !sm.IndexExists(index) {
		sm.log.Info(fmt.Sprint("create index at time: ", d.Format(time.RFC3339)))
		err := sm.CreateIndex(index, sd[0].GetMapping())
		if err != nil {
			sm.log.Err(err.Error())
		}
	}
	i := 0
	bir := make([]*elastic.BulkIndexRequest, l)
	for _, s := range sd {
		d, err := s.GetJsonBody()
		if err != nil {
			continue
		}
		bir[i] = elastic.NewBulkIndexRequest().Index(index).Id(s.GetID()).Doc(d)
		i++
	}

	bulkRequest := sm.search.Bulk()
	for _, b := range bir {
		bulkRequest.Add(b)
	}
	_, err := bulkRequest.Do(sm.ctx)
	if err != nil {
		return err
	}
	sm.log.Debug("save to new search")
	return nil
}

func (sm *searchModel) PutByDate(sd dao.ESDao, t time.Time) error {
	index := sm.esDI.GetIndexByTime(t, sd.GetType())
	if !sm.IndexExists(index) {
		sm.CreateIndex(index, sd.GetMapping())
	}
	body, err := sd.GetJsonBody()
	if err != nil {
		return err
	}

	_, err = sm.search.Index().
		Index(index).
		Id(sd.GetID()).
		BodyJson(body).Do(sm.ctx)

	if err != nil {
		return err
	}

	_, err = sm.search.Flush().Index(index).Do(sm.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (sm *searchModel) IndexExists(index string) bool {
	sm.Lock()
	if exist, ok := sm.indexMap[index]; ok {
		return exist
	}
	exists, err := sm.search.IndexExists(index).Do(sm.ctx)
	if err != nil {
		return false
	}
	if !exists {
		return false
	}
	sm.indexMap[index] = true
	sm.Unlock()
	return true
}

func (sm *searchModel) CreateIndex(index, mapping string) error {
	createIndex, err := sm.search.CreateIndex(index).Body(mapping).Do(sm.ctx)
	if err != nil {
		return err
	}
	if !createIndex.Acknowledged {
		return errors.New("cant create index.")
	}
	sm.Lock()
	sm.indexMap[index] = true
	sm.Unlock()
	return nil
}
