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
	Key   string   `json:"key"`
	Perm  gcp.Perm `json:"perm"`
	X     float64  `json:"x"`
	Y     float64  `json:"y"`
	W     float64  `json:"width"`
	H     float64  `json:"height"`
	Alpha float64  `json:"alpha"`

	sto gcp.Storage
	ctx context.Context
}

func (ele *ImageElement) Generate(p pdf.PDF) error {
	if ele.sto == nil {
		return errors.New("not set storage")
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

	p.ImageReader(imgH2, ele.X, ele.Y, ele.W, ele.H, ele.Alpha)
	return nil
}
