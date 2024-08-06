package main

import (
	"context"
	"database/sql"
	"time"

	rpc "github.com/werydude/graven-server/internal/rpc"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(
	ctx context.Context,
	logger runtime.Logger,
	db *sql.DB,
	nk runtime.NakamaModule,
	initializer runtime.Initializer,
) error {
	var initStart time.Time = time.Now()

	//go:generate echo "Building RPC: Health Check"
	var err error = initializer.RegisterRpc("healthcheck", rpc.RpcHealthcheck)
	if err != nil {
		return err
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())
	return nil
}
