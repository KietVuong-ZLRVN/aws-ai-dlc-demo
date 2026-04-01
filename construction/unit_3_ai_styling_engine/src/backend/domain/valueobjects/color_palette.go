package valueobjects

import "errors"

type ColorPalette struct {
	Preferred []Color
	Excluded  []Color
}

func NewColorPalette(preferred, excluded []Color) (ColorPalette, error) {
	prefSet := make(map[Color]bool, len(preferred))
	for _, c := range preferred {
		prefSet[c] = true
	}
	for _, c := range excluded {
		if prefSet[c] {
			return ColorPalette{}, errors.New("a color cannot appear in both preferred and excluded lists")
		}
	}
	return ColorPalette{Preferred: preferred, Excluded: excluded}, nil
}
