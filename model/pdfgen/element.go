package pdfgen

import (
	"context"
	"errors"

	"github.com/94peter/sterna/export/pdf"
	"github.com/94peter/sterna/gcp"
	"github.com/94peter/sterna/util"

	"github.com/signintech/gopdf"
)

type PdfElement interface {
	Generate(p pdf.PDF) error
}

type Element struct {
	Type string `json:"type"`
}

var (
	alignMap = map[string]int{
		"left":   pdf.AlignLeft,
		"center": pdf.AlignCenter,
		"right":  pdf.AlignRight,
	}
)

type TextElemnt struct {
	*Element
	X        float64   `json:"x"`
	Y        float64   `json:"y"`
	Field    string    `json:"field"`
	Split    string    `json:"split"`
	Value    string    `json:"value"`
	Font     string    `json:"font"`
	FontSize int       `json:"fontSize"`
	Color    pdf.Color `json:"color"`
	Align    string    `json:"align"`
}

func (ele *TextElemnt) Generate(p pdf.PDF) error {
	if ele.X > 0 {
		p.SetX(ele.X)
	}
	if ele.Y > 0 {
		p.SetY(ele.Y)
	}

	p.Text(ele.Text(), pdf.TextStyle{
		Font:     ele.Font,
		FontSize: ele.FontSize,
		Color:    ele.Color,
	}, alignMap[ele.Align])
	return nil
}

func (ele *TextElemnt) Text() string {
	return util.StrAppend(ele.Field, ele.Split, ele.Value)
}

type BrElemnt struct {
	*Element
	H float64
}

func (ele *BrElemnt) Generate(p pdf.PDF) error {
	p.Br(ele.H)
	return nil
}

type LineElemnt struct {
	*Element
	Width float64 `json:"width"`
	Style string  `json:"style"`
}

func (ele *LineElemnt) Generate(p pdf.PDF) error {
	switch ele.Style {
	case "dashed":
		p.NewDashLine(ele.Width)
	default:
		p.NewLine(ele.Width)
	}
	return nil
}

type ImageElement struct {
	*Element
	Field    string    `json:"field"`
	Font     string    `json:"font"`
	FontSize int       `json:"fontSize"`
	Color    pdf.Color `json:"color"`
	Key      string    `json:"key"`
	Perm     gcp.Perm  `json:"perm"`
	X        float64   `json:"x"`
	Y        float64   `json:"y"`
	W        float64   `json:"width"`
	H        float64   `json:"height"`
	Alpha    float64   `json:"alpha"`
	Spacing  float64   `json:"spacing"`
	DSKey    string    `json:"dsKey"`

	ds  interface{}
	sto gcp.Storage
	ctx context.Context
}

func (ele *ImageElement) Generate(p pdf.PDF) error {
	if ele.sto == nil {
		return errors.New("not set storage")
	}
	if ele.DSKey == "" {
		return errors.New("dsKey not set")
	}
	if data, ok := ele.ds.(map[string]interface{}); ok {
		imgMap, ok := data[ele.DSKey].(map[string]string)
		if !ok {
			return errors.New("data source not found")
		}
		ele.Key, ok = imgMap[ele.Key]
		if !ok {
			return errors.New("image key not found: " + ele.Key)
		}
	}
	textw, _ := p.MeasureTextWidth(ele.Field)
	if ele.X == 0 {
		ele.X = p.GetX()
	}
	if ele.Y == 0 {
		ele.Y = p.GetY()
	}
	reader, err := ele.sto.OpenFile(ele.ctx, ele.Key, ele.Perm)
	if err != nil {
		return err
	}
	imgH2, err := gopdf.ImageHolderByReader(reader)
	if err != nil {
		return err
	}
	if ele.Alpha <= 0 {
		ele.Alpha = 1.0
	}
	p.ImageReader(imgH2, ele.X+textw, ele.Y, ele.W, ele.H, ele.Alpha)
	moveY := (ele.H / 2) - float64(ele.FontSize/2)
	p.SetY(ele.Y + moveY)
	p.SetX(ele.X)
	p.Text(ele.Field, pdf.TextStyle{
		Font:     ele.Font,
		FontSize: ele.FontSize,
		Color:    ele.Color,
	}, pdf.AlignLeft)
	p.SetX(ele.X + textw + ele.W + ele.Spacing)
	p.SetY(ele.Y)
	return nil
}

type TableElement struct {
	*Element
	Title     []tableTitleRow `json:"title"`
	DSKey     string          `json:"dsKey"`
	X         float64         `json:"x"`
	Y         float64         `json:"y"`
	RowHeight float64         `json:"rowHeight"`
	Alpha     float64         `json:"alpha"`

	ds interface{}
}

type tableTitleRow struct {
	Name     string    `json:"name"`
	Key      string    `json:"key"`
	Width    float64   `json:"width"`
	Font     string    `json:"font"`
	FontSize int       `json:"fontSize"`
	Color    pdf.Color `json:"color"`
	Align    string    `json:"align"`
}

func (ele *TableElement) Generate(p pdf.PDF) error {
	if ele.DSKey == "" {
		return errors.New("dsKey not set")
	}
	if ele.X > 0 {
		p.SetX(ele.X)
	}
	if ele.Y > 0 {
		p.SetY(ele.Y)
	}
	for _, h := range ele.Title {
		p.RectFillDrawColor(h.Name, pdf.TextBlockStyle{
			TextStyle: pdf.TextStyle{
				Font:     h.Font,
				FontSize: h.FontSize,
				Color:    h.Color,
			},
			BackGround: pdf.ColorWhite,
		}, h.Width, ele.RowHeight, alignMap[h.Align], pdf.ValignMiddle)
	}
	p.Br(ele.RowHeight)
	p.SetX(ele.X)
	if data, ok := ele.ds.(map[string]interface{}); ok {
		rows, ok := data[ele.DSKey].([]map[string]string)
		if !ok {
			return errors.New("data source not found")
		}
		var value string
		for _, r := range rows {
			for _, h := range ele.Title {
				value, ok = r[h.Key]
				if !ok {
					value = ""
				}
				p.RectFillDrawColor(value, pdf.TextBlockStyle{
					TextStyle: pdf.TextStyle{
						Font:     h.Font,
						FontSize: h.FontSize,
						Color:    h.Color,
					},
					BackGround: pdf.ColorWhite,
				}, h.Width, ele.RowHeight, alignMap[h.Align], pdf.ValignMiddle)
			}
			p.Br(ele.RowHeight)
			p.SetX(ele.X)
		}
	}
	return nil
}
