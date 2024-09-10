package match

import "github.com/werydude/graven-server/internal/cards"

type RevealData struct {
	InstanceCards []cards.InstanceCard `json:"instance_cards"`
	Reveal        bool                 `json:"reveal"`
}
