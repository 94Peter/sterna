package pdf

type Color struct {
	R, G, B, A uint8
}

var (
	ColorTableLine = Color{
		R: 0,
		G: 0,
		B: 0,
	}

	ColorWhite = Color{
		R: 255,
		G: 255,
		B: 255,
	}

	ColorGray = Color{
		R: 211,
		G: 211,
		B: 211,
	}

	ColorBlack = ColorTableLine
)
