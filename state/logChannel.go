package state

import (
	"context"
	"log/slog"
	"time"
)

type GuildLogChannel struct {
	ID int64 `bun:",pk,autoincrement"`

	GuildID   uint64 `bun:",nullzero,notnull"`
	ChannelID uint64 `bun:",nullzero,notnull"`

	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

func (s *State) GuildLogChannelGet(ctx context.Context, guildId uint64) (uint64, error) {
	var out GuildLogChannel
	err := s.db.NewSelect().Model(&out).Where("guild_id = ?", guildId).Scan(ctx)
	if err != nil {
		return 0, err
	}

	return out.ChannelID, nil
}

func (s *State) GuildLogChannelSet(ctx context.Context, guildId, channelId uint64) error {
	s.logger.Info("setting log channel", slog.Any("guildId", guildId), slog.Any("channelId", channelId))

	_, err := s.db.NewInsert().Model(&GuildLogChannel{
		GuildID:   guildId,
		ChannelID: channelId,
	}).Exec(ctx)
	return err
}
