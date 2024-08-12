package match

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"golang.org/x/exp/maps"
)

type OpCode int

const (
	Connected OpCode = 0
	Ready     OpCode = 1 << iota
	Playing
	Draw
)

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

type DrawnCard struct {
	CardId Card   `json:"CardID"`
	NodeId []byte `json:"NodeID"`
}

func OnDraw(mState *MatchState, message_ptr *runtime.MatchData, logger *runtime.Logger, dispatcher *runtime.MatchDispatcher, data []byte) *MatchState {
	message := *message_ptr
	player := mState.Players[message.GetUserId()]
	drawn := DrawnCard{
		player.Data.DrawCard(),
		data,
	}

	mState.Players[message.GetUserId()] = player

	if enc_data, err := json.Marshal(drawn); err == nil {
		(*dispatcher).BroadcastMessage(int64(Draw), enc_data, nil, player.Presence, true)
		(*logger).Warn("BROADCASTED!")
	} else {
		(*logger).Error("%s", err)
	}
	return mState

}
