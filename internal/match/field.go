package match

type FieldData struct {
	Deck   []Card `json:"deck"`
	Grave  []Card `json:"grave"`
	Hand   []Card `json:"hand"`
	Survey []Card `json:"survey"`
	Effect []Card `json:"effect"`
}

func (field *FieldData) DrawCard() Card {
	var drawn Card
	field.Deck, field.Hand, drawn = PopMove[Card](field.Deck, field.Hand)

	return drawn
}

func PopMove[T comparable](zoneA []T, zoneB []T) ([]T, []T, T) {
	pop := zoneA[0]
	zoneB = append(zoneB, pop)
	zoneA = append(zoneA[:0], zoneA[1:]...)
	return zoneA, zoneB, pop
}
