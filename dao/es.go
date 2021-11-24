package dao

import "errors"

type ESDao interface {
	GetType() string
	GetJsonBody() (string, error)
	GetID() string
	GetMapping() string
}

type SearchDaoList []ESDao

func (sdl SearchDaoList) Add(sd ESDao) error {
	if len(sdl) == 0 {
		sdl = append(sdl, sd)
		return nil
	}
	if sd.GetType() != sdl[0].GetType() {
		return errors.New("dao type error")
	}
	sdl = append(sdl, sd)
	return nil
}

type index struct {
	Mappings *mappings `json:"mappings"`
}

func NewEsIndex() *index {
	return &index{
		Mappings: &mappings{
			Properties: make(map[string]*property),
		},
	}
}

type mappings struct {
	Properties map[string]*property `json:"properties"`
}

type property struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

const (
	PropFormatEmpty = ""
	PropFormatTime  = "epoch_second"
)

func (p *index) AddProperty(key string, _type string, _format string) {
	if p.Mappings == nil {
		p.Mappings = &mappings{
			Properties: make(map[string]*property),
		}
	}

	pr := &property{
		Type: _type,
	}
	if _format != "" {
		pr.Format = _format
	}
	p.Mappings.Properties[key] = pr
}
