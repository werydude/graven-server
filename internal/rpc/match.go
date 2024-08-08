package rpc

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type ParamPayload struct {
	Params map[string]interface{} `json:"params"`
}

func RpcCreateStandardMatch(
	ctx context.Context,
	logger runtime.Logger,
	db *sql.DB,
	nk runtime.NakamaModule,
	payload string,

) (string, error) {
	var param_payload ParamPayload
	if payload_err := json.Unmarshal([]byte(payload), &param_payload); payload_err != nil {
		logger.Error("Error marshalling response type to JSON: %v", payload_err)
		return "", runtime.NewError("Cannon marshal type", 13)
	}

	var (
		matchID string
		err     error
	)
	matchID, err = nk.MatchCreate(ctx, "standard_match", param_payload.Params)
	logger.Debug("Match created! (matchID: %s)", matchID)

	return matchID, err
}
