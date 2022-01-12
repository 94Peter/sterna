package pdfgen

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/94peter/sterna/export/pdf"
	"github.com/94peter/sterna/gcp"

	"github.com/stretchr/testify/assert"
)

type Data struct {
	Name string
	Data *Data
}

func Test_genPdfRender(t *testing.T) {
	data := map[string]interface{}{
		"Vatnumber":     "250778098",
		"Establishment": "108年9月12日",
		"Register":      "臺灣臺北地方法院登記處-登記簿第172冊第57頁第3639號",
		"Approval":      "台內團字第1080060869號",
		"Address":       "臺北市中山區撫順街35號3樓",
		"Contact":       "(02)2599-1258  eafatku@gmail.com",
		"Title":         "社團法人淡江大學卓越校友會",
		"No":            "NO.110-001",
		"Name":          "陳彥睿",
		"Data": []map[string]string{
			{"no": "1", "item": "入會費", "amount": "1,000"},
			{"no": "2", "item": "常年會費", "amount": "500"},
		},
		"Image": map[string]string{
			"chairman": "img1.png",
			"officer":  "img1.png",
			"big":      "img2.png",
		},
		"PrintDate": "中華民國 110 年 12 月 1 日",
	}
	ctx := context.Background()
	storage := &gcp.GcpConf{
		CredentialsFile: "kalimas-279302.json",
		Bucket:          "tmp-kalimasi-storage",
	}
	conf := NewPdfGenConf(ctx, data, storage)
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
