package match

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

type OpCode int

const (
	Connected OpCode = 0
	Ready     OpCode = 1 << iota
	Playing
	Draw
	Move
	Reveal
	EndMatch
)

func DecodeOpCode(mState *MatchState, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, message runtime.MatchData) *MatchState {
	(*logger).Info("Received %v from %v (OpCode: %v", string(message.GetData()), message.GetUserId(), message.GetOpCode())
	switch OpCode(message.GetOpCode()) {
	case Connected:
		(*logger).Warn("Connected OPCODE")
	case Ready:
		OnReady(mState, &message, logger, dispatcher)
	case Draw:

		mState = OnDraw(mState, &message, logger, dispatcher, message.GetData())

		if CheckDeck(mState, message.GetUserId()) {
			loser := message.GetUserId()
			mState.Loser = &(loser)
		}
	case Move:
		(*logger).Warn("Running OnMove")
		mState = OnMove(mState, &message, logger, dispatcher, message.GetData())

	case Reveal:
		mState = OnReveal(mState, &message, logger, dispatcher, message.GetData())
	}
	return mState
}

func CheckDeck(mState *MatchState, user_id string) bool {
	return len(mState.Players[user_id].Data.Deck.Contents) == 0
}

func CheckReady(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger) (bool, *MatchState) {
	message := *message_ptr

	player := mState.Players[message.GetUserId()]
	(*logger).Warn("Message Data: %b | int: %d | cmp: %t", message.GetData(), message.GetData()[0], message.GetData()[0] == 1)

	if message.GetData()[0] == 1 {
		player.State = Ready
	} else {
		player.State = Connected
	}

	mState.Players[message.GetUserId()] = player

	(*logger).Warn("State: %d", player.State)
	(*logger).Warn("mState: %d", mState.Players[message.GetUserId()].State)

	if len(mState.Players) != int(MAX_PLAYERS) {
		(*logger).Warn("Failed Count")
		return false, mState
	}
	(*logger).Warn("Passed count")
	//(*logger).Warn("Players: %+v", mState.Players)
	for _, player := range mState.Players {
		if player.State != Ready {
			(*logger).Warn("Failed ready")
			return false, mState
		}
	}
	(*logger).Warn("success!")
	return true, mState
}

func OnReady(mState *MatchState, message *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher) {
	ready, newState := CheckReady(mState, message, logger)
	(*logger).Warn("%+v", newState)
	(*logger).Warn(fmt.Sprintf("%t", ready))
	if ready {

		keys := statesAsBtyes(mState.Players, logger)
		(*logger).Warn(string(keys))
		if len(keys) > 0 {
			(*dispatcher).BroadcastMessage(int64(Playing), keys, nil, nil, true)
		}
	}
}

func statesAsBtyes(playerStates map[string]PlayerState, logger *runtime.Logger) []byte {

	out, err := json.Marshal(playerStates)
	if err != nil {
		(*logger).Error("Error marshalling response type to JSON: %v", err)
		return make([]byte, 0)
	}
	return out
}

func OnDraw(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]
	(*logger).Warn("DATA: %s", data)
	var msg MoveCommandMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		(*logger).Warn("Failed to unmashal; %s (%s)->(%+v)", err, data, msg)
	} else {
		(*logger).Warn("unmarshal Success! %+v", msg)
	}
	drawn := player.Data.DrawCard(msg.B.Tag)

	(*logger).Warn("%+v", player)
	mState.Players[message.GetUserId()] = player
	msg.Card = drawn

	game_data := DeckCounter{
		msg, len(player.Data.Deck.Contents),
	}
	if enc_data, err := json.Marshal(game_data); err == nil {
		(*dispatcher).BroadcastMessage(int64(Draw), enc_data, nil, player.Presence, true)
		(*logger).Warn("broadcasted!")
	} else {
		(*logger).Error("%s", err)
	}
	return mState

}

func OnMove(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]

	var msg MoveCommandMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		(*logger).Warn("Failed to unmashal; %s (%s)->(%+v)", err, data, msg)
	} else {
		if msg.B != nil {
			(*logger).Warn("unmarshal Success! (%+v, %+v, %+v)", msg.A, *msg.B, msg.Card)
		} else {
			(*logger).Warn("unmarshal Success! (%+v, NIL, %+v)", msg.A)
		}
	}

	cmd, err := msg.CreateCommand(mState, logger)
	if err != nil {
		(*logger).Error("Failed to create command; %s (%+v)->(%+v)", err, msg, cmd)
	} else {
		(*logger).Warn("Created command! (%v)->(%+v); (%+v =Card=> %+v)", msg, cmd, *cmd.ZoneA, *cmd.ZoneB)
	}

	newZoneA, newZoneB, err := cmd.run(logger)
	if err != nil {
		(*logger).Warn("Failed command: %s", err)
	}
	fieldA, err := (*mState).GetPlayerFieldData(msg.A.Player, nil)
	fieldB, err := (*mState).GetPlayerFieldData(msg.B.Player, nil)

	*fieldA.ZoneFromTag(msg.A.Tag) = *newZoneA
	*fieldB.ZoneFromTag(msg.B.Tag) = *newZoneB

	(*logger).Warn("FIELDS -- fieldA: %+v, fieldB: %+v", fieldA, fieldB)

	(*logger).Warn("NEW ZONES (OnMove) --> ZoneA: %+v, ZoneB: %+v", cmd.ZoneA, cmd.ZoneB)
	game_data := DeckCounter{
		msg, len(player.Data.Deck.Contents),
	}
	if enc_data, err := json.Marshal(game_data); err == nil {
		(*dispatcher).BroadcastMessage(int64(Move), enc_data, nil, player.Presence, true)
		(*logger).Warn("broadcasted!")
	} else {
		(*logger).Error("%s", err)
	}

	return mState

}

func OnReveal(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]
	var reveal = string(data) == "true"
	reveal_data := RevealData{
		InstanceCards: player.Data.Hand.Contents,
		Reveal:        reveal,
	}
	if enc_data, err := json.Marshal(reveal_data); err == nil {
		(*dispatcher).BroadcastMessage(int64(Reveal), enc_data, nil, player.Presence, true)
		(*logger).Warn("broadcasted!")
	} else {
		(*logger).Error("%s", err)
	}

	return mState
}
