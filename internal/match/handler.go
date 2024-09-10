package match

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/werydude/graven-server/internal/cards"
)

const DECK_SIZE uint8 = 32
const MAX_PLAYERS uint8 = 2

type PlayerState struct {
	Presence runtime.Presence `json:"presence"`
	Data     FieldData        `json:"data"`
	State    OpCode           `json:"opcode"`
}

func NewPlayerState(p_presence runtime.Presence, p_instances string, logger runtime.Logger) PlayerState {
	deck := Zone{Contents: cards.DecodeInstances(p_instances, logger)}

	grave := NewZone(DECK_SIZE)
	hand := NewZone(DECK_SIZE)
	survey := NewZone(DECK_SIZE)
	effect := NewZone(DECK_SIZE)

	return PlayerState{
		p_presence,
		FieldData{
			deck,
			grave,
			hand,
			survey,
			effect,
		},
		Connected,
	}

}

type DeckCounter struct {
	Data      interface{} `json:"data"`
	DeckCount int         `json:"deck_count"`
}

type MatchState struct {
	Players map[string]PlayerState `json:"players"`
	Loser   *string
}

func (m_state MatchState) GetPlayerFieldData(player_string string, logger *runtime.Logger) (*FieldData, error) {
	var player = m_state.Players[player_string]
	// TODO: Error handle
	return &player.Data, nil
}

type Match struct{}

func NewMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		Players: make(map[string]PlayerState),
	}

	tickRate := 30
	var label string = fmt.Sprintf("%s", params["label"])
	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {

	mState, _ := state.(*MatchState)
	if len(mState.Players) >= 2 {
		return state, false, "Room is full!"
	}
	instances := ""
	if ival, exists := metadata["instances"]; exists {
		instances = ival
	}

	mState.Players[presence.GetUserId()] = NewPlayerState(presence, instances, logger)
	// player_presences := make([]runtime.Presence, 0, len(mState.Players))
	//
	// for player := range mState.Players {
	// 	player_presences = append(player_presences, mState.Players[player].Presence)
	// }
	//
	// dispatcher.BroadcastMessage(int64(Connected), make([]byte, 0, 0), nil, presence, true)

	logger.Warn("BROADCASTED JOIN (%+v)", presence)
	return state, true, fmt.Sprintf("%+v", mState.Players[presence.GetUserId()].Data.Deck)
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		_, exists := mState.Players[p.GetUserId()]
		if exists {
			continue
		}
		mState.Players[p.GetUserId()] = NewPlayerState(p, "", logger)
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		delete(mState.Players, p.GetUserId())
	}

	if len(mState.Players) == 0 {
		return nil
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	for _, player := range mState.Players {
		presence := player.Presence
		if false {
			logger.Info("Presence %v named %v", presence.GetUserId(), presence.GetUsername())
		}
	}

	for _, message := range messages {
		mState = DecodeOpCode(mState, &logger, &dispatcher, message)
	}
	if mState.Loser != nil {
		dispatcher.BroadcastMessage(int64(EndMatch), []byte(*mState.Loser), nil, nil, true)
		return nil
	}
	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
