package match

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
	cards "github.com/werydude/graven-server/internal/cards"
)

const DECK_SIZE uint8 = 32
const MAX_PLAYERS uint8 = 2

type PlayerState struct {
	Presence runtime.Presence
	Data     FieldData
}

type FieldData struct {
	Deck   []string `json:"deck"`
	Grave  []string `json:"grave"`
	Hand   []string `json:"hand"`
	Survey []string `json:"survey"`
	Effect []string `json:"effect"`
}

func NewPlayerState(p_presence runtime.Presence, p_deckcode string, logger runtime.Logger) PlayerState {
	var deck []string
	decoded_deck, dde := cards.DecodeDeckCode(p_deckcode, logger)
	if dde.Err != nil || dde.Err == nil {
		deck = append(deck, decoded_deck...)
	} else {
		logger.Warn("%s", dde)
	}
	return PlayerState{
		p_presence,
		FieldData{
			deck,
			make([]string, DECK_SIZE),
			make([]string, DECK_SIZE),
			make([]string, DECK_SIZE),
			make([]string, DECK_SIZE),
		},
	}

}

type MatchState struct {
	Players map[string]PlayerState `json:"players"`
}

type Match struct{}

func NewMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		Players: make(map[string]PlayerState),
	}

	tickRate := 1
	label := ""

	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {

	mState, _ := state.(*MatchState)
	if len(mState.Players) >= 2 {
		return state, false, "Room is full!"
	}
	deck_code := ""
	if val, exists := metadata["deck_code"]; exists {
		deck_code = val
	}
	mState.Players[presence.GetUserId()] = NewPlayerState(presence, deck_code, logger)
	logger.Warn("%s", mState.Players[presence.GetUserId()].Data.Deck)
	return state, true, fmt.Sprintf("%s", mState.Players[presence.GetUserId()].Data.Deck)
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
		logger.Info("Presence %v named %v", presence.GetUserId(), presence.GetUsername())
	}

	for _, message := range messages {
		logger.Info("Received %v from %v", string(message.GetData()), message.GetUserId())
		reliable := true
		dispatcher.BroadcastMessage(1, message.GetData(), []runtime.Presence{message}, nil, reliable)
	}

	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	message := "Server shutting down in " + strconv.Itoa(graceSeconds) + " seconds."
	reliable := true
	dispatcher.BroadcastMessage(2, []byte(message), []runtime.Presence{}, nil, reliable)

	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
