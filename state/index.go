package state

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type State struct {
	db     *bun.DB
	logger *slog.Logger
}

func New(ctx context.Context, logger *slog.Logger) (*State, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, "onyx.db")
	if err != nil {
		return nil, eris.Wrap(err, "failed to open sqlite database")
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	state := &State{
		db,
		logger,
	}

	err = state.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *State) Close() error {
	return s.db.Close()
}

func (s *State) Migrate(ctx context.Context) error {
	s.logger.Info("migrating database")

	s.logger.Debug("creating guild_log_channel table")
	_, err := s.db.NewCreateTable().IfNotExists().Model((*GuildLogChannel)(nil)).Exec(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to create guild_log_channel table")
	}

	s.logger.Debug("creating guild_message table")
	_, err = s.db.NewCreateTable().IfNotExists().Model((*GuildMessage)(nil)).Exec(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to create guild_message table")
	}

	return nil
}
