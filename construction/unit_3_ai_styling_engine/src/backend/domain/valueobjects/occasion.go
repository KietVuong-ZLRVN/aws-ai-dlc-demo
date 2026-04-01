package valueobjects

type Occasion string

const (
	OccasionCasual  Occasion = "casual"
	OccasionFormal  Occasion = "formal"
	OccasionOutdoor Occasion = "outdoor"
	OccasionBeach   Occasion = "beach"
	OccasionOffice  Occasion = "office"
	OccasionParty   Occasion = "party"
)

var ValidOccasions = []Occasion{
	OccasionCasual, OccasionFormal, OccasionOutdoor,
	OccasionBeach, OccasionOffice, OccasionParty,
}
