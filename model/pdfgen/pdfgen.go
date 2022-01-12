package pdfgen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"

	"github.com/94peter/sterna/export"
	"github.com/94peter/sterna/export/pdf"
	"github.com/94peter/sterna/gcp"
)

func NewPdfGenConf(ctx context.Context, data interface{}, storage gcp.Storage) *pdfGenConf {
	return &pdfGenConf{
		RenderData: data,
		sto:        storage,
		ctx:        ctx,
	}
}

type pdfGenConf struct {
	RenderData  interface{}
	Elements    []PdfElement      `json:"-"`
	RawElements []json.RawMessage `json:"elements"`

	sto gcp.Storage
	ctx context.Context
}

func (conf *pdfGenConf) AddElemnts(ele PdfElement) {
	conf.Elements = append(conf.Elements, ele)
}

func (f *pdfGenConf) MarshalJSON() ([]byte, error) {
	type conf pdfGenConf
	if f.Elements != nil {
		for _, v := range f.Elements {
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			f.RawElements = append(f.RawElements, b)
		}
	}
	return json.Marshal((*conf)(f))
}

func (f *pdfGenConf) UnmarshalJSON(b []byte) error {
	if f.RenderData != nil {
		t, err := template.New("tpl").Parse(string(b))
		if err != nil {
			return err
		}
		var tpl bytes.Buffer
		if err = t.Execute(&tpl, f.RenderData); err != nil {
			return err
		}
		b = tpl.Bytes()
	}
	type conf pdfGenConf
	err := json.Unmarshal(b, (*conf)(f))
	if err != nil {
		return err
	}

	for _, raw := range f.RawElements {
		var v Element
		err = json.Unmarshal(raw, &v)
		if err != nil {
			return err
		}
		var i PdfElement
		switch v.Type {
		case "text":
			i = &TextElemnt{}
		case "br":
			i = &BrElemnt{}
		case "line":
			i = &LineElemnt{}
		case "image":
			i = &ImageElement{sto: f.sto, ctx: f.ctx, ds: f.RenderData}
		case "table":
			i = &TableElement{ds: f.RenderData}
		default:
			return errors.New("unknown element type: " + v.Type)
		}
		err = json.Unmarshal(raw, i)
		if err != nil {
			return err
		}
		f.Elements = append(f.Elements, i)
	}
	return nil
}

func NewPdfGen(mypdf pdf.PDF, genConf *pdfGenConf) export.PDF {
	mypdf.AddPage()
	return &pdfGen{
		mypdf:   mypdf,
		genConf: genConf,
	}
}

type pdfGen struct {
	mypdf   pdf.PDF
	genConf *pdfGenConf
}

func (gen *pdfGen) Pdf(w io.Writer) error {
	var err error
	for _, e := range gen.genConf.Elements {
		err = e.Generate(gen.mypdf)
		if err != nil {
			return err
		}
	}
	return gen.mypdf.Write(w)
}
