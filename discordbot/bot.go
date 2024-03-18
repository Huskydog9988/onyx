package discordbot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/huskydog9988/onyx/state"
	"github.com/rotisserie/eris"
)

var token = ""

type Onyx struct {
	client bot.Client
	state  *state.State
}

func New(ctx context.Context, logger *slog.Logger) *Onyx {

	onyx := &Onyx{}

	// create a new state
	stateManager, err := state.New(ctx, logger)
	if err != nil {
		logger.Error("failed to create state", slog.Any("error", err))
		panic(err)
	}

	// set the state
	onyx.state = stateManager

	client, err := disgo.New(token,
		// set gateway options
		bot.WithGatewayConfigOpts(
			// set enabled intents
			gateway.WithIntents(
				// https://discord.com/developers/docs/topics/gateway#list-of-intents
				gateway.IntentGuilds,
				gateway.IntentGuildMembers,
				gateway.IntentGuildModeration,
				gateway.IntentGuildInvites,
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
				// gateway.IntentDirectMessages,
			),
		),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			logger.Info("ready", slog.String("username", e.User.Username))
		}),
		// add event listeners
		bot.WithEventListenerFunc(func(e *events.MessageCreate) {
			// ignore bot messages
			if e.Message.Author.Bot {
				return
			}

			// save message
			onyx.state.GuildMessageCreate(ctx, state.GuildMessage{
				MessageID: uint64(e.MessageID),
				Content:   e.Message.Content,
			})
		}),
		bot.WithEventListenerFunc(func(e *events.GuildMessageUpdate) {
			// get old message content
			msg, err := onyx.state.GuildMessageGet(ctx, uint64(e.MessageID))
			if err != nil {
				logger.Error("failed to get message", slog.Any("error", err))
				return
			}

			// save new content
			err = onyx.state.GuildMessageUpdate(ctx, state.GuildMessage{
				MessageID: uint64(e.MessageID),
				Content:   e.Message.Content,
			})
			if err != nil {
				logger.Error("failed to update message", slog.Any("error", err))
				return
			}

			logger.Info("message updated", slog.String("new", e.Message.Content), slog.String("old", msg.Content))
		}),
		bot.WithEventListenerFunc(func(e *events.GuildMessageDelete) {
			// get old message content
			msg, err := onyx.state.GuildMessageGet(ctx, uint64(e.MessageID))
			if err != nil {
				logger.Error("failed to get message", slog.Any("error", err))
				return
			}

			// delete message
			err = onyx.state.GuildMessageDelete(ctx, uint64(e.MessageID))
			if err != nil {
				logger.Error("failed to delete message", slog.Any("error", err))
				return
			}

			logger.Info("message deleted", slog.String("content", msg.Content))
		}),
	)
	if err != nil {
		msg := "failed to create disgo client"
		err := eris.Wrap(err, msg)
		logger.Error(msg, slog.Any("error", err))
		panic(err)
	}
	// connect to the gateway
	if err = client.OpenGateway(ctx); err != nil {
		msg := "failed to open gateway"
		err := eris.Wrap(err, msg)
		logger.Error(msg, slog.Any("error", err))
		panic(err)
	}

	// set the client
	onyx.client = client

	return onyx
}

func (o *Onyx) Close(ctx context.Context) {
	// close the client
	o.client.Close(ctx)

	// close the state
	o.state.Close()
}
