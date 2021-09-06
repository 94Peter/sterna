package pdf

var (
	alignMap = map[string]int{
		"left":   AlignLeft,
		"right":  AlignRight,
		"center": AlignCenter,
	}
)

type TextStyle struct {
	Font     string
	FontSize int
	Color    Color
	Style    string
}

type TextBlockStyle struct {
	TextStyle
	BackGround Color
	W, H       float64
	TextAlign  string
}

func (tbs *TextBlockStyle) GetAlign() int {
	align, ok := alignMap[tbs.TextAlign]
	if !ok {
		return AlignLeft
	}
	return align
}
