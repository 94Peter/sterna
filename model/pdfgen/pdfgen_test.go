package pdfgen

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/94peter/sterna/export/pdf"

	"github.com/stretchr/testify/assert"
)

type Data struct {
	Name string
	Data *Data
}

func Test_genPdfRender(t *testing.T) {
	data := &Data{
		Name: "render",
		Data: &Data{
			Name: "subrender",
		},
	}
	ctx := context.Background()
	conf := NewPdfGenConf(ctx, data, nil)
	f, err := os.Open("test2.json")
	assert.Nil(t, err)
	err = json.NewDecoder(f).Decode(conf)
	assert.Nil(t, err)

	pdfFile, err := os.Create("test2.pdf")
	assert.Nil(t, err)
	mypdf := pdf.GetA4PDF(&pdf.PdfConf{
		Font: pdf.PdfFont{
			"tw-m": "/Users/peter/goproject/remora-go/rsrc/pdf/TW-Medium.ttf",
			"tw-r": "/Users/peter/goproject/remora-go/rsrc/pdf/TW-Regular.ttf",
		}}, 20, 20, 20, 20)
	gen := NewPdfGen(mypdf, conf)
	err = gen.Pdf(pdfFile)
	assert.Nil(t, err)
}
func Test_genPDF(t *testing.T) {
	ctx := context.Background()
	conf := NewPdfGenConf(ctx, nil, nil)
	f, err := os.Open("test1.json")
	assert.Nil(t, err)
	err = json.NewDecoder(f).Decode(conf)
	assert.Nil(t, err)

	pdfFile, err := os.Create("test.pdf")
	assert.Nil(t, err)
	mypdf := pdf.GetA4PDF(&pdf.PdfConf{
		Font: pdf.PdfFont{
			"tw-m": "/Users/peter/goproject/remora-go/rsrc/pdf/TW-Medium.ttf",
			"tw-r": "/Users/peter/goproject/remora-go/rsrc/pdf/TW-Regular.ttf",
		}}, 20, 20, 20, 20)
	gen := NewPdfGen(mypdf, conf)
	err = gen.Pdf(pdfFile)
	assert.Nil(t, err)
}

func Test_jsonMarshal(t *testing.T) {
	ctx := context.Background()
	conf := NewPdfGenConf(ctx, nil, nil)
	conf.AddElemnts(&TextElemnt{
		Element: &Element{
			Type: "text",
		},
		X:        20,
		Field:    "aaaa",
		Value:    "bbbbb",
		Font:     "tw-r",
		FontSize: 12,
		Color:    pdf.ColorBlack,
		Align:    "left",
	})
	f, err := os.Create("test.json")
	assert.Nil(t, err)
	err = json.NewEncoder(f).Encode(conf)
	assert.Nil(t, err)
}
