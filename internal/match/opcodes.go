package match

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/werydude/graven-server/internal/cards"
	"golang.org/x/exp/maps"
)

type OpCode int

const (
	Connected OpCode = 0
	Ready     OpCode = 1 << iota
	Playing
	Draw
	Move
	GetDeck
)

func DecodeOpCode(mState *MatchState, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, message runtime.MatchData) *MatchState {
	(*logger).Info("Received %v from %v (OpCode: %v, need: %v)", string(message.GetData()), message.GetUserId(), message.GetOpCode(), Ready)
	switch OpCode(message.GetOpCode()) {
	case Ready:
		OnReady(mState, &message, logger, dispatcher)
	case Draw:
		mState = OnDraw(mState, &message, logger, dispatcher, message.GetData())
	case Move:
		(*logger).Warn("Running OnMove")
		mState = OnMove(mState, &message, logger, dispatcher, message.GetData())
	case GetDeck:
		mState = OnGetDeck(mState, &message, logger, dispatcher, message.GetData())
	}
	return mState
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
	(*logger).Warn("Players: %+v", mState.Players)
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

		keys := keysAsBtyes(mState.Players, logger)
		(*logger).Warn(string(keys))
		if len(keys) > 0 {
			(*dispatcher).BroadcastMessage(int64(Playing), keys, nil, nil, true)
		}
	}
}

func keysAsBtyes(playerStates map[string]PlayerState, logger *runtime.Logger) []byte {
	players := make(map[string]string, len(maps.Keys(playerStates)))
	for id, state := range playerStates {
		players[id] = state.Presence.GetUsername()
	}

	out, err := json.Marshal(players)
	if err != nil {
		(*logger).Error("Error marshalling response type to JSON: %v", err)
		return make([]byte, 0)
	}
	return out
}

func OnDraw(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]

	var cmd MoveCommand
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		(*logger).Warn("Failed to unmashal; %s (%s)->(%+v)", err, data, cmd)
	} else {
		(*logger).Warn("unmarshal Success! %+v", cmd)
	}
	drawn := player.Data.DrawCard(cmd.TagB)
	(*logger).Warn("%+v", player)
	mState.Players[message.GetUserId()] = player
	cmd.Card = drawn
	if enc_data, err := json.Marshal(cmd); err == nil {
		(*dispatcher).BroadcastMessage(int64(Draw), enc_data, nil, player.Presence, true)
		(*logger).Warn("BROADCASTED!")
	} else {
		(*logger).Error("%s", err)
	}
	return mState

}

func OnMove(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]

	var cmd MoveCommand
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		(*logger).Warn("Failed to unmashal; %s (%s)->(%+v)", err, data, cmd)
	}

	new_data, err := cmd.run(&player.Data, logger)
	if err != nil {
		(*logger).Warn("Failed command: %s", err)
	} else {
		player.Data = *new_data
		mState.Players[message.GetUserId()] = player

		(*dispatcher).BroadcastMessage(int64(Move), data, nil, player.Presence, true)
	}
	return mState

}

func OnGetDeck(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]

	deck_code, err := cards.EncodeDeckCode(player.Data.Deck, *logger)

	if err.Err != nil {
		(*logger).Warn("Failed encode: %s", err)
	} else {
		(*dispatcher).BroadcastMessage(int64(GetDeck), deck_code, []runtime.Presence{player.Presence}, player.Presence, true)
	}
	return mState

}
