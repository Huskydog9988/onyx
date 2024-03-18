package discordbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/huskydog9988/onyx/state"
	"github.com/rotisserie/eris"
)

var token = ""
var testGuildID = snowflake.ID(492075852071174144)

type Onyx struct {
	client bot.Client
	state  *state.State
	logger *slog.Logger

	avatarURL string
}

func New(ctx context.Context, logger *slog.Logger) *Onyx {

	onyx := &Onyx{
		// default avatar, replaced later
		avatarURL: "https://i.ytimg.com/vi/RGS8A3j81fY/maxresdefault.jpg",
	}

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
			gateway.WithAutoReconnect(true),

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

		bot.WithLogger(logger.WithGroup("disgo")),
		bot.WithEventListeners(onyx.commandHandler()),
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
				AuthorID:  uint64(e.Message.Author.ID),
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

			logChannelID, err := onyx.state.GuildLogChannelGet(ctx, uint64(e.GuildID))
			if err != nil {
				logger.Error("failed to get log channel", slog.Any("error", err))
				return
			}

			// member, ok := e.Client().Caches().Member(e.GuildID, snowflake.ID(msg.AuthorID))
			// if !ok {
			// 	logger.Error("failed to get user from cache")
			// 	return
			// }

			user, err := e.Client().Rest().GetUser(snowflake.ID(msg.AuthorID))
			if err != nil {
				logger.Error("failed to get user", slog.Any("error", err))
				return
			}

			// e.Client().Rest().CreateMessage(snowflake.ID(logChannelID), discord.NewMessageCreateBuilder().SetContent("Message deleted: "+msg.Content).Build())

			embed := newOnyxLogEmbed()
			embed.SetAuthor(*user)
			embed.SetId("Message", e.MessageID)
			embed.SetColor(OnyxLogEmbedColorWarn)
			embed.AddField("Content", msg.Content, false)
			embed.AddDateField()
			embed.SetFooter(onyx.avatarURL)
			embed.SetDescription(fmt.Sprintf("Message deleted in <#%d>", e.ChannelID))

			_, err = e.Client().Rest().CreateMessage(snowflake.ID(logChannelID),
				discord.NewMessageCreateBuilder().SetEmbeds(embed.Build()).Build())
			if err != nil {
				logger.Error("failed to send message", slog.Any("error", err))
				return
			}

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

	if err = handler.SyncCommands(client, commands, []snowflake.ID{testGuildID}); err != nil {
		msg := "error while syncing commands"
		err := eris.Wrap(err, msg)
		logger.ErrorContext(ctx, msg, slog.Any("error", err))
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
