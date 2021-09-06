package pdf

import (
	"sterna/util"
)

type TableStyle interface {
	Header() TextBlockStyle
	Data() TextBlockStyle
}

type columnData struct {
	TextAry  []string
	StyleAry []TextBlockStyle
}

func (p *pdfImpl) DrawColumn(w, h float64, color Color, rectType string) {
	p.SetFillColor(color.R, color.G, color.B)
	p.RectFromUpperLeftWithStyle(p.GetX(), p.GetY(), w, h, rectType)
}

// 文字區塊4個數字
func (p *pdfImpl) TextValuesAry(rows columnData, valign int) {

	ox, x := p.GetX(), 0.0

	if ox < p.leftMargin {
		ox = p.leftMargin
	}
	x = ox

	oy, y := p.GetY(), p.GetY()

	maxFontSize := 8.0

	for i := 0; i < len(rows.TextAry); i++ {
		p.SetX(x)
		text := rows.TextAry[i]
		tStyle := rows.StyleAry[i]
		x = p.tableText(text, maxFontSize, tStyle.Color, x, y, tStyle.GetAlign(), valign, tStyle.W, tStyle.W)
	}
	y = oy + maxFontSize
	p.SetY(y)
	p.SetX(ox)
}

func (p *pdfImpl) tableText(text string, floatFontSize float64, color Color, x, y float64, align, valign int, w, h float64) (endX float64) {
	ox := x
	p.SetFillColor(color.R, color.G, color.B)
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

	if valign == ValignMiddle {
		y = y + (h / 2) - (floatFontSize / 2)
	} else if valign == ValignBottom {
		y = y + h - floatFontSize
	}
	p.SetY(y)
	p.Cell(nil, text)
	endX = ox + w
	return
}

type NormalTable struct {
	header   []TableHeaderColumn
	widthAry []float64
	rows     [][]*TableColumn
}

func GetNormalTable() *NormalTable {
	return &NormalTable{}
}

func (nti *NormalTable) AddHeader(thc TableHeaderColumn) {
	if len(thc.Sub) == 0 {
		nti.widthAry = append(nti.widthAry, thc.Main.Width)
	} else {
		for _, s := range thc.Sub {
			nti.widthAry = append(nti.widthAry, s.Width)
		}
	}

	nti.header = append(nti.header, thc)
}

func (nti *NormalTable) AddRow(r []*TableColumn) {
	nti.rows = append(nti.rows, r)
}

func (nti *NormalTable) Draw(p PDF, style TableStyle, app ...AddPagePipe) {
	if len(nti.rows) > 0 && len(nti.widthAry) != len(nti.rows[0]) {
		panic("table setting error")
	}

	ox, x := p.GetX(), 0.0

	if ox < p.GetLeftMargin() {
		ox = p.GetLeftMargin()
	}
	x = ox
	p.SetX(x)

	//oy, y := pdf.GetY(), pdf.GetY()
	y := p.GetY()
	p.SetY(y)

	maxHeight := 0.0
	for _, h := range nti.header {
		h.draw(p, style)
		p.SetY(y)
		height := h.Main.Height
		if len(h.Sub) > 0 {
			height += h.Sub[0].Height
		}
		if height > maxHeight {
			maxHeight = height
		}
	}
	p.Br(maxHeight)
	rowY := 0.0
	for _, r := range nti.rows {
		maxHeight = 0
		i := 0
		rowY = p.GetY()
		for _, h := range r {
			h.Height = h.Height / float64(len(h.Text))
			x = p.GetX()
			if maxHeight < h.Height {
				maxHeight = h.Height
			}
			j := 0.0
			aIndex := 0
			for _, text := range h.Text {
				if h.Align[aIndex] == AlignRight {
					text = util.StrAppend(text, "  ")
				}
				p.SetY(p.GetY() + h.Height*j)
				p.SetX(x)
				p.RectFillDrawColor(text, style.Data(), nti.widthAry[i], h.Height, h.Align[aIndex], ValignMiddle)
				j++
				aIndex++
			}
			p.SetY(rowY)
			i++
		}
		p.Br(maxHeight)
		if p.GetHeight()-p.GetY() < p.GetBottomMargin()*3 {
			p.Br(10)
			p.AddPagePipe(app...)
		}
	}

}

func (nti *NormalTable) DrawWithPosition(p PDF, style TableStyle, px, py float64) {
	if len(nti.rows) > 0 && len(nti.widthAry) != len(nti.rows[0]) {
		panic("table setting error")
	}
	ox, x := p.GetX(), 0.0

	if px < p.GetLeftMargin() {
		px = p.GetLeftMargin()
	}
	x = px
	p.SetX(x)

	//oy, y := pdf.GetY(), pdf.GetY()
	oy := p.GetY()
	p.SetY(oy)

	y := py
	maxHeight := 0.0
	for _, h := range nti.header {
		h.draw(p, style)
		p.SetY(y)
		height := h.Main.Height
		if len(h.Sub) > 0 {
			height += h.Sub[0].Height
		}
		if height > maxHeight {
			maxHeight = height
		}
	}
	p.Br(maxHeight)
	rowY := 0.0
	for _, r := range nti.rows {
		maxHeight = 0
		i := 0
		rowY = p.GetY()
		p.SetX(px)
		for _, h := range r {
			h.Height = h.Height / float64(len(h.Text))
			x = p.GetX()
			if maxHeight < h.Height {
				maxHeight = h.Height
			}
			j := 0.0
			aIndex := 0

			for _, text := range h.Text {
				if h.Align[aIndex] == AlignRight {
					text = util.StrAppend(text, "  ")
				}
				p.SetY(p.GetY() + h.Height*j)
				p.SetX(x)
				p.RectFillDrawColor(text, style.Data(), nti.widthAry[i], h.Height, h.Align[aIndex], ValignMiddle)
				j++
				aIndex++
			}
			p.SetY(rowY)
			i++
		}
		p.Br(maxHeight)
		if p.GetHeight()-p.GetY() < p.GetBottomMargin()*1.5 {
			p.AddPage()
		}
	}
	p.SetX(ox)
	p.SetY(oy)
}

type TableColumn struct {
	Text   []string
	Width  float64
	Height float64
	Align  []int
}

func GetTableColumn(w, h float64, align []int, text ...string) *TableColumn {
	tl := len(text)
	if len(align) < tl {
		oldAlign := align[0]
		align = make([]int, len(text))
		for i := 0; i < tl; i++ {
			align[i] = oldAlign
		}
	}
	return &TableColumn{
		Text:   text,
		Width:  w,
		Height: h,
		Align:  align,
	}
}

type TableHeaderColumn struct {
	Main TableColumn
	Sub  []TableColumn
}

func (thc *TableHeaderColumn) AddSub(text string, width, height float64) {
	thc.Sub = append(thc.Sub, TableColumn{Text: []string{text}, Width: width, Height: height})
}

func (nti *TableHeaderColumn) draw(p PDF, style TableStyle) {
	ox, x := p.GetX(), 0.0

	if ox < p.GetLeftMargin() {
		ox = p.GetLeftMargin()
	}

	x = ox
	p.SetX(x)
	p.SetY(p.GetY())
	p.RectFillDrawColor(nti.Main.Text[0], style.Header(), nti.Main.Width, nti.Main.Height, AlignCenter, ValignMiddle)
	if len(nti.Sub) == 0 {
		return
	}
	p.SetY(p.GetY() + nti.Main.Height)
	maxSubHeight := 0.0
	for _, t := range nti.Sub {
		p.SetX(x)
		p.RectFillDrawColor(t.Text[0], style.Header(), t.Width, t.Height, AlignCenter, ValignMiddle)
		x = x + t.Width
		if maxSubHeight < t.Height {
			maxSubHeight = t.Height
		}
	}
	p.SetY(p.GetY() + maxSubHeight)

}
