package discordbot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
)

var commands = []discord.ApplicationCommandCreate{}

// commandHandler handles all the commands
func (o *Onyx) commandHandler() *handler.Mux {
	r := handler.New()

	// register the settings router
	r.Route("/settings", o.settingsRouter())

	r.NotFound(func(e *events.InteractionCreate) error {
		o.logger.Warn("interaction not found", slog.Any("interactionID", e.Interaction.ID()), slog.Any("createdAt", e.Interaction.CreatedAt()))

		return nil
	})

	r.Error(func(e *events.InteractionCreate, err error) {
		o.logger.Error("error handling interaction", slog.Any("err", err))
	})

	return r
}

// settingsRouter handles the settings command
func (o *Onyx) settingsRouter() func(r handler.Router) {
	// add the settings command to the list of commands
	commands = append(commands, discord.SlashCommandCreate{
		Name:        "settings",
		Description: "Change the settings of the bot",
		Options: []discord.ApplicationCommandOption{

			// audit log channel subcommand
			discord.ApplicationCommandOptionSubCommand{
				Name:        "log-channel",
				Description: "Sets the channel to log events to",
				Options: []discord.ApplicationCommandOption{

					// channel option
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel to log events to",
						Required:    true,
					},
				},
			},
		},
	})

	return func(r handler.Router) {
		r.Command("/log-channel", func(e *handler.CommandEvent) error {
			e.DeferCreateMessage(true)

			channel := e.SlashCommandInteractionData().Channel("channel")

			o.state.GuildLogChannelSet(context.TODO(), uint64(*e.GuildID()), uint64(channel.ID))

			return nil
		})

	}
}
