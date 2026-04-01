package valueobjects

type StyleDirection string

const (
	StyleMinimalist StyleDirection = "minimalist"
	StyleBold       StyleDirection = "bold"
	StyleClassic    StyleDirection = "classic"
	StyleBohemian   StyleDirection = "bohemian"
)

var ValidStyleDirections = []StyleDirection{
	StyleMinimalist, StyleBold, StyleClassic, StyleBohemian,
}
