package pdf

import (
	"io"
	"os"
	"strings"

	"github.com/signintech/gopdf"
)

type PdfFont map[string]string

type PdfConf struct {
	Font PdfFont
}

type AddPagePipe interface {
	Before(p PDF)
	After(p PDF)
}

type PDF interface {
	WriteToFile(filepath string) error
	Write(w io.Writer) error
	AddPagePipe(pp ...AddPagePipe)
	AddPage()
	Text(text string, style TextStyle, align int)
	Br(h float64)

	GetX() float64
	GetY() float64
	SetX(x float64)
	SetY(y float64)
	GetBottomHeight() float64
	GetBottomMargin() float64
	GetLeftMargin() float64
	GetRightMargin() float64
	GetHeight() float64
	NewLine(width float64)
	NewDashLine(width float64)
	LineWithPosition(width float64, x1, y1, x2, y2 float64)
	RectFillDrawColor(text string, style TextBlockStyle, w, h float64, align, valign int)
	ImageReader(imageByte io.Reader, x, y, w, h, alpha float64) error
}

type pdfImpl struct {
	*gopdf.GoPdf
	width, height                                    float64
	leftMargin, topMargin, rightMargin, bottomMargin float64
	page                                             uint8
}

func GetA4PDF(conf *PdfConf, leftMargin, rightMargin, topMargin, bottomMargin float64) PDF {
	gpdf := gopdf.GoPdf{}
	width, height := 595.28, 841.89
	gpdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: width, H: height}}) //595.28, 841.89 = A4
	var err error
	for key, value := range conf.Font {
		err = gpdf.AddTTFFont(key, value)
		if err != nil {
			panic(err)
		}
	}
	gpdf.SetLeftMargin(leftMargin)
	gpdf.SetTopMargin(topMargin)
	return &pdfImpl{
		GoPdf:        &gpdf,
		width:        width,
		height:       height,
		leftMargin:   leftMargin,
		rightMargin:  rightMargin,
		topMargin:    topMargin,
		bottomMargin: bottomMargin,
	}
}

func GetA4HPDF(conf *PdfConf, leftMargin, rightMargin, topMargin, bottomMargin float64) PDF {
	gpdf := gopdf.GoPdf{}
	height, width := 595.28, 841.89
	gpdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: width, H: height}}) //595.28, 841.89 = A4
	var err error
	for key, value := range conf.Font {
		err = gpdf.AddTTFFont(key, value)
		if err != nil {
			panic(err)
		}
	}
	gpdf.SetLeftMargin(leftMargin)
	gpdf.SetTopMargin(topMargin)
	return &pdfImpl{
		GoPdf:        &gpdf,
		width:        width,
		height:       height,
		leftMargin:   leftMargin,
		rightMargin:  rightMargin,
		topMargin:    topMargin,
		bottomMargin: bottomMargin,
	}
}

const (
	ValignTop    = 1
	ValignMiddle = 2
	ValignBottom = 3
	AlignLeft    = 4
	AlignCenter  = 5
	AlignRight   = 6
)

func (p *pdfImpl) WriteToFile(filepath string) error {
	return p.WritePdf(filepath)
}

func (p *pdfImpl) Write(w io.Writer) error {
	return p.GoPdf.Write(w)
}
func (p *pdfImpl) GetPage() uint8 {
	return p.page
}
func (p *pdfImpl) AddPagePipe(pp ...AddPagePipe) {
	for _, t := range pp {
		t.Before(p)
	}
	p.page++
	p.AddPage()
	for _, t := range pp {
		t.After(p)
	}
}

func (p *pdfImpl) Br(h float64) {
	p.GoPdf.Br(h)
	p.SetX(p.leftMargin)
}

func (p *pdfImpl) GetWidth() float64 {
	return p.width - p.leftMargin - p.rightMargin
}

func (p *pdfImpl) GetBottomMargin() float64 {
	return p.bottomMargin
}

func (p *pdfImpl) GetLeftMargin() float64 {
	return p.leftMargin
}

func (p *pdfImpl) GetRightMargin() float64 {
	return p.rightMargin
}

func (p *pdfImpl) GetBottomHeight() float64 {
	return p.height - p.bottomMargin
}

func (p *pdfImpl) GetHeight() float64 {
	return p.height
}

func (p *pdfImpl) NewLine(width float64) {
	p.SetLineType("solid")
	p.LineWithPosition(width, p.leftMargin, p.GetY(), p.width-p.rightMargin, p.GetY())
}

func (p *pdfImpl) NewDashLine(width float64) {
	p.SetLineType("dashed")
	p.LineWithPosition(width, p.leftMargin, p.GetY(), p.width-p.rightMargin, p.GetY())
}

func (p *pdfImpl) LineWithPosition(width float64, x1, y1, x2, y2 float64) {
	p.SetLineWidth(width)
	p.Line(x1, y1, x2, y2)
}

func (p *pdfImpl) TextWithPosition(text string, style TextStyle, x, y float64) {
	p.SetFont(style.Font, "", style.FontSize)
	textw, _ := p.MeasureTextWidth(text)
	rightLimit := p.width - p.rightMargin - textw
	if x < p.leftMargin {
		x = p.leftMargin
	} else if x > rightLimit {
		x = rightLimit
	}
	oy := p.GetY()
	ox := p.GetX()
	p.SetX(x)
	p.SetY(y)

	color := style.Color
	p.SetTextColor(color.R, color.G, color.B)
	p.SetFillColor(color.R, color.G, color.B)

	p.Cell(nil, text)
	p.SetX(ox)
	p.SetY(oy)
}

func (p *pdfImpl) Text(text string, style TextStyle, align int) {
	p.SetFont(style.Font, "", style.FontSize)
	color := style.Color
	p.SetTextColor(color.R, color.G, color.B)
	p.SetFillColor(color.R, color.G, color.B)
	ox := p.GetX()
	if ox < p.leftMargin {
		ox = p.leftMargin
	}
	x := ox
	textw, _ := p.MeasureTextWidth(text)
	switch align {
	case AlignCenter:
		x = (p.width / 2) - (textw / 2)
	case AlignRight:
		x = p.width - textw - p.rightMargin
	}
	p.SetX(x)
	p.Cell(nil, text)
	p.SetX(ox + textw)
}

func (p *pdfImpl) TwoColumnText(text1, text2 string, style TextStyle) {
	p.SetFont(style.Font, "", style.FontSize)
	color := style.Color
	p.SetTextColor(color.R, color.G, color.B)
	p.SetX(p.leftMargin)
	p.Cell(nil, text1)
	p.SetX(p.width/2 + p.leftMargin)
	p.Cell(nil, text2)
}

func (p *pdfImpl) ImageReader(imageByte io.Reader, x, y, w, h, alpha float64) error {
	//use image holder by io.Reader
	imgH2, err := gopdf.ImageHolderByReader(imageByte)
	if err != nil {
		return err
	}
	var rect *gopdf.Rect
	if w > 0 && h > 0 {
		rect = &gopdf.Rect{W: w, H: h}
	}
	transparency, err := gopdf.NewTransparency(alpha, "")
	if err != nil {
		return err
	}

	imOpts := gopdf.ImageOptions{
		X:            x,
		Y:            y,
		Transparency: &transparency,
		Rect:         rect,
	}
	return p.ImageByHolderWithOptions(imgH2, imOpts)
}

func (p *pdfImpl) Image(imagePath string) {
	//use image holder by io.Reader
	file, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	imgH2, err := gopdf.ImageHolderByReader(file)
	if err != nil {
		panic(err)
	}
	p.ImageByHolder(imgH2, p.leftMargin, p.GetY(), nil)
}

func (p *pdfImpl) RectDrawColor(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
) {
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "D")
}

func (p *pdfImpl) RectDrawColorWithPosition(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
	x, y float64,
) {
	ox, oy := p.GetX(), p.GetY()
	p.SetX(x)
	p.SetY(y)
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "D")
	p.SetX(ox)
	p.SetY(oy)
}

func (p *pdfImpl) RectFillDrawColor(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
) {
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "FD")
}

func (p *pdfImpl) RectFillDrawColorWithPosition(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
	x, y float64,
) {
	ox, oy := p.GetX(), p.GetY()
	p.SetX(x)
	p.SetY(y)
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "FD")
	p.SetX(ox)
	p.SetY(oy)
}

func (p *pdfImpl) RectFillColorWithPosition(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
	x, y float64,
) {
	ox, oy := p.GetX(), p.GetY()
	p.SetX(x)
	p.SetY(y)
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "F")
	p.SetX(ox)
	p.SetY(oy)
}

func (p *pdfImpl) RectFillColor(text string,
	style TextBlockStyle,
	w, h float64,
	align, valign int,
) {
	p.rectColorText(text, style.Font, style.Style, style.FontSize, style.Color, w, h, style.BackGround, align, valign, "F")
}

func (p *pdfImpl) rectColorText(text string,
	font string,
	style string,
	fontSize int,
	textColor Color,
	w, h float64,
	color Color,
	align, valign int,
	rectType string,
) {
	p.SetLineWidth(0.1)
	p.SetFont(font, style, fontSize)
	p.SetFillColor(color.R, color.G, color.B) //setup fill color
	ox, x := p.GetX(), 0.0

	if ox < p.leftMargin {
		ox = p.leftMargin
	}
	x = ox
	p.RectFromUpperLeftWithStyle(x, p.GetY(), w, h, rectType)
	p.SetFillColor(0, 0, 0)

	s := strings.Split(text, "\n")
	text = s[0]
	for _, t := range s {
		if len(t) > len(text) {
			text = t
		}
	}
	if align == AlignCenter {
		textw, _ := p.MeasureTextWidth(text)
		x = x + (w / 2) - (textw / 2)
	} else if align == AlignRight {
		textw, _ := p.MeasureTextWidth(text)
		x = x + w - textw
	} else {
		x = x + 5
	}

	p.SetX(x)
	oy, y := p.GetY(), 0.0
	if valign == ValignMiddle {
		y = oy + (h / 2) - (float64(fontSize) * float64(len(s)) / 2)
	} else if valign == ValignBottom {
		y = oy + h - float64(fontSize)*float64(len(s))
	}
	p.SetY(y)

	p.SetTextColor(textColor.R, textColor.G, textColor.B)

	i := 0.0
	for _, t := range s {
		p.SetY(y + float64(fontSize)*i)
		p.SetX(x)
		p.Cell(nil, t)
		i++
	}

	p.SetY(oy)
	p.SetX(ox + w)

}
