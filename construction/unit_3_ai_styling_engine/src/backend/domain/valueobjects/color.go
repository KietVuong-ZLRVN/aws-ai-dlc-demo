package valueobjects

type Color string

const (
	ColorBlack  Color = "black"
	ColorWhite  Color = "white"
	ColorNavy   Color = "navy"
	ColorBeige  Color = "beige"
	ColorRed    Color = "red"
	ColorGreen  Color = "green"
	ColorPastel Color = "pastel"
)

var ValidColors = []Color{
	ColorBlack, ColorWhite, ColorNavy, ColorBeige,
	ColorRed, ColorGreen, ColorPastel,
}
