package main

import (
	"context"
	"database/sql"
	"time"

	match "github.com/werydude/graven-server/internal/match"
	rpc "github.com/werydude/graven-server/internal/rpc"

	"github.com/heroiclabs/nakama-common/runtime"
)

type initBundle struct {
	ctx         context.Context
	logger      runtime.Logger
	db          *sql.DB
	nk          runtime.NakamaModule
	initializer runtime.Initializer
}

func InitModule(
	ctx context.Context,
	logger runtime.Logger,
	db *sql.DB,
	nk runtime.NakamaModule,
	initializer runtime.Initializer,
) error {
	var initStart time.Time = time.Now()
	bundle := initBundle{
		ctx,
		logger,
		db,
		nk,
		initializer,
	}
	rpcErr := registerRpcs(bundle)
	if rpcErr != nil {
		logger.Error("[RegisterRpcs] error: ", rpcErr.Error())
		return rpcErr
	}

	matchesErr := registerMatches(bundle)
	if matchesErr != nil {
		logger.Error("[RegisterMatches] error: ", matchesErr.Error())
		return matchesErr
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())
	return nil
}

func registerRpcs(bundle initBundle) error {
	//go:generate echo "Building RPC: Health Check"
	if err := bundle.initializer.RegisterRpc("healthcheck", rpc.RpcHealthcheck); err != nil {
		return err
	}

	if err := bundle.initializer.RegisterRpc("create_standard_match", rpc.RpcCreateStandardMatch); err != nil {
		return err
	}

	return nil
}

func registerMatches(bundle initBundle) error {

	bundle.logger.Info("Hello Multiplayer!")
	//go:generate echo "Building standard_match register"
	var err error = bundle.initializer.RegisterMatch("standard_match", match.NewMatch)
	return err
}
