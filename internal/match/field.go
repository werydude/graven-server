package match

import (
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/werydude/graven-server/internal/cards"
)

type Zone struct {
	Contents []cards.InstanceCard
}

func NewZone(size uint8) Zone {
	return Zone{
		Contents: make([]cards.InstanceCard, 0, size),
	}
}

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
	zone := field.ZoneFromTag(tag)
	field.Deck.Contents, (*zone).Contents, drawn = PopMove[cards.InstanceCard](field.Deck.Contents, (*zone).Contents, 0)
	return drawn
}

func PopMove[T any](zoneA []T, zoneB []T, idx int) ([]T, []T, T) {
	pop := zoneA[idx]
	zoneB = append(zoneB, pop)
	zoneA = append(zoneA[:idx], zoneA[idx+1:]...)
	return zoneA, zoneB, pop
}

func MoveCard(zoneA_ptr *Zone, zoneB_ptr *Zone, card cards.InstanceCard, logger *runtime.Logger) (*Zone, *Zone, error) {

	if logger != nil {
		(*logger).Warn("ZONE A: %+v \n ZONE B: %+v", zoneA_ptr, zoneB_ptr)
	}
	var err error

	contentsA, contentsB, err := MoveItem((*zoneA_ptr).Contents, (*zoneB_ptr).Contents, card, logger)
	zoneA_ptr.Contents = contentsA
	zoneB_ptr.Contents = contentsB

	if logger != nil {
		(*logger).Warn("ZONE A: %+v \n ZONE B: %+v", zoneA_ptr, zoneB_ptr)
	}
	return zoneA_ptr, zoneB_ptr, err
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

type MoveCommandMessage struct {
	A    MoveCommandParticipant  `json:"originPlayer"`
	B    *MoveCommandParticipant `json:"targetPlayer"`
	Card cards.InstanceCard      `json:"card"`
}
type MoveCommandParticipant struct {
	Player string  `json:"player"`
	Tag    ZoneTag `json:"tag"`
}

type MoveCommand struct {
	ZoneA *Zone
	ZoneB *Zone
	Card  cards.InstanceCard
}

func (cmd MoveCommand) run(logger *runtime.Logger) (*Zone, *Zone, error) {
	(*logger).Warn("RUNNING COMMAND. ZoneA: %+v, ZoneB: %+v", *cmd.ZoneA, *cmd.ZoneB)
	newZoneA, newZoneB, err := MoveCard(cmd.ZoneA, cmd.ZoneB, cmd.Card, logger)
	(*logger).Warn("NEW ZONES (run) --> ZoneA: %+v, ZoneB: %+v", cmd.ZoneA, cmd.ZoneB)
	return newZoneA, newZoneB, err
}

// TODO: Error handle
func (msg MoveCommandMessage) CreateCommand(mState *MatchState, logger *runtime.Logger) (MoveCommand, error) {
	if msg.B == nil {
		msg.B = &msg.A
	}
	var A, B = msg.A, *msg.B

	var fieldA_ptr, _ = (*mState).GetPlayerFieldData(A.Player, logger)
	var fieldB_ptr, _ = (*mState).GetPlayerFieldData(B.Player, logger)

	var ZoneA, ZoneB = (*fieldA_ptr).ZoneFromTag(A.Tag), (*fieldB_ptr).ZoneFromTag(B.Tag)

	return MoveCommand{
		ZoneA,
		ZoneB,
		msg.Card,
	}, nil
}

func MoveItem(contentsA []cards.InstanceCard, contentsB []cards.InstanceCard, item cards.InstanceCard, logger *runtime.Logger) ([]cards.InstanceCard, []cards.InstanceCard, error) {
	var node_idx int = -1

	for i, c := range contentsA {
		if c.NodeId == item.NodeId {
			node_idx = i
			(*logger).Warn("Found in A!")
			break
		}
	}
	if node_idx < 0 {
		return contentsA, contentsB, errors.New("item not found")
	}
	contentsA, contentsB, _ = PopMove(contentsA, contentsB, node_idx)
	return contentsA, contentsB, nil
}
