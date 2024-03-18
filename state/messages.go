package state

import (
	"context"
	"time"
)

type GuildMessage struct {
	MessageID uint64 `bun:",pk"`

	Content string `bun:",nullzero,notnull"`

	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

func (s *State) GuildMessageGet(ctx context.Context, messageId uint64) (GuildMessage, error) {
	var out GuildMessage
	err := s.db.NewSelect().Model(&out).Where("message_id = ?", messageId).Scan(ctx)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (s *State) GuildMessageCreate(ctx context.Context, message GuildMessage) error {
	_, err := s.db.NewInsert().Model(&message).Exec(ctx)

	return err
}

func (s *State) GuildMessageUpdate(ctx context.Context, message GuildMessage) error {
	// only update the content
	_, err := s.db.NewUpdate().Model(&message).WherePK().Column("content").Exec(ctx)

	return err
}

func (s *State) GuildMessageDelete(ctx context.Context, messageId uint64) error {
	_, err := s.db.NewDelete().Model(&GuildMessage{}).Where("message_id = ?", messageId).Exec(ctx)

	return err
}
