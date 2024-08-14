package match

import (
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/werydude/graven-server/internal/cards"
)

type Zone []cards.InstanceCard
type ZoneTag int

const (
	DeckTag ZoneTag = iota
	GraveTag
	HandTag
	SurveyTag
	EffectTag
)

type FieldData struct {
	Deck   Zone `json:"deck"`
	Grave  Zone `json:"grave"`
	Hand   Zone `json:"hand"`
	Survey Zone `json:"survey"`
	Effect Zone `json:"effect"`
}

func (field *FieldData) DrawCard(tag ZoneTag) cards.InstanceCard {
	var drawn cards.InstanceCard
	new_zone := field.ZoneFromTag(tag)
	field.Deck, *new_zone, drawn = PopMove[cards.InstanceCard](field.Deck, *new_zone, 0)
	return drawn
}

func PopMove[T any](zoneA []T, zoneB []T, idx int) ([]T, []T, T) {
	pop := zoneA[idx]
	zoneB = append(zoneB, pop)
	zoneA = append(zoneA[:idx], zoneA[idx+1:]...)
	return zoneA, zoneB, pop
}

func (field *FieldData) MoveZone(tagA ZoneTag, tagB ZoneTag, card cards.InstanceCard, logger *runtime.Logger) (*FieldData, error) {
	zoneA_ptr, zoneB_ptr := field.ZoneFromTag(tagA), field.ZoneFromTag(tagB)
	if logger != nil {
		(*logger).Warn("ZONE A: %+v \n ZONE B: %+v", zoneA_ptr, zoneB_ptr)
		(*logger).Warn("TAG A: %+v \n TAG B: %+v", tagA, tagB)
	}
	var err error
	*zoneA_ptr, *zoneB_ptr, err = MoveItem(*zoneA_ptr, *zoneB_ptr, card, logger)
	return field, err
}

func (field *FieldData) ZoneFromTag(tag ZoneTag) *Zone {
	switch tag {
	case DeckTag:
		return &field.Deck
	case GraveTag:
		return &field.Grave
	case HandTag:
		return &field.Hand
	case SurveyTag:
		return &field.Survey
	case EffectTag:
		return &field.Effect
	}
	return nil
}

type MoveCommand struct {
	TagA ZoneTag            `json:"tagA"`
	TagB ZoneTag            `json:"tagB"`
	Card cards.InstanceCard `json:"card"`
}

func (cmd MoveCommand) run(field *FieldData, logger *runtime.Logger) (*FieldData, error) {
	return field.MoveZone(cmd.TagA, cmd.TagB, cmd.Card, logger)
}

func MoveItem(zoneA []cards.InstanceCard, zoneB []cards.InstanceCard, item cards.InstanceCard, logger *runtime.Logger) ([]cards.InstanceCard, []cards.InstanceCard, error) {
	var node_idx int = -1
	var card_idx int = -1
	if logger != nil {
		(*logger).Warn("ZONE A: %+v", zoneA)
	}
	for i, c := range zoneA {
		if c.NodeId == item.NodeId {
			node_idx = i
			break
		}
		if c.CardId == item.CardId {
			card_idx = i
		}
	}
	if card_idx > 0 && node_idx == -1 {
		node_idx = card_idx
	}
	if node_idx < 0 {
		return zoneA, zoneB, errors.New("item not found")
	}
	zoneA, zoneB, _ = PopMove(zoneA, zoneB, node_idx)
	return zoneA, zoneB, nil
}
