package search

import "encoding/json"

type IndexBuilder interface {
	AddProperty(string, PropType, PropFormat) IndexBuilder
	String() string
}

type index struct {
	Mappings *mappings `json:"mappings"`
}

func NewIndexBuilder() IndexBuilder {
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
	Type   PropType   `json:"type"`
	Format PropFormat `json:"format,omitempty"`
}

type PropFormat string

const (
	PropFormatEmpty = PropFormat("")
	PropFormatTime  = PropFormat("epoch_second")
)

type PropType string

const (
	PropType_Keyword = PropType("keyword")
	PropType_Date    = PropType("date")
	PropType_Double  = PropType("double")
	PropType_Int     = PropType("integer")
)

func (p *index) AddProperty(key string, typ PropType, format PropFormat) IndexBuilder {
	if p.Mappings == nil {
		p.Mappings = &mappings{
			Properties: make(map[string]*property),
		}
	}

	pr := &property{
		Type:   typ,
		Format: format,
	}
	p.Mappings.Properties[key] = pr
	return p
}

func (p *index) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(data)
}
