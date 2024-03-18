package discordbot

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

const (
	OnyxLogEmbedColorRed    = 0xD90202
	OnyxLogEmbedColorPurple = 0x822aed
	OnyxLogEmbedColorPink   = 0xe62aed
)

type OnyxLogEmbed struct {
	embed *discord.EmbedBuilder

	ids map[string]snowflake.ID

	time time.Time
}

func newOnyxLogEmbed() OnyxLogEmbed {
	return OnyxLogEmbed{
		embed: discord.NewEmbedBuilder(),
		ids:   make(map[string]snowflake.ID),
		time:  time.Now(),
	}
}

// sets the custom date field
func (o *OnyxLogEmbed) AddDateField() {
	o.AddField("Date", fmt.Sprintf("<t:%d:F>", o.time.Unix()), false)
}

// sets the channel field based on the message
func (o *OnyxLogEmbed) AddChannelMessageField(msg discord.Message) {
	o.AddField("Channel", fmt.Sprintf("<#%d>\n[Go To Message](%s)", msg.ChannelID, msg.JumpURL()), false)
}

// sets the channel field based on the channelID
func (o *OnyxLogEmbed) AddChannelField(channelID snowflake.ID) {
	o.AddField("Channel", fmt.Sprintf("<#%d>", channelID), false)
}

func (o *OnyxLogEmbed) AddDifferanceFields(now, previous string) {
	o.AddField("Now", now, false)
	o.AddField("Previous", previous, false)

}

func (o *OnyxLogEmbed) AddField(name, value string, inline bool) {
	o.embed.AddField(name, value, inline)
}

// sets the author field based on the user
func (o *OnyxLogEmbed) SetAuthor(author discord.User) {
	o.SetId("User", author.ID)
	o.embed.SetAuthor(author.Username, "", *author.AvatarURL())
}

func (o *OnyxLogEmbed) SetDescription(description string) {
	o.embed.SetDescription(description)
}

func (o *OnyxLogEmbed) Build() discord.Embed {
	o.buildIds()

	return o.embed.Build()
}

func (o *OnyxLogEmbed) SetId(key string, id snowflake.ID) {
	o.ids[key] = id
}

func (o *OnyxLogEmbed) SetColor(color int) {
	o.embed.SetColor(color)
}

func (o *OnyxLogEmbed) buildIds() {
	if len(o.ids) == 0 {
		// don't need to build anything if there are no special props
		return
	}

	props := ""
	for k, v := range o.ids {
		props += k + " = " + v.String() + "\n"
	}
	props = fmt.Sprintf("```INI\n%s```", props)

	o.embed.AddField("ID", props, false)
}

func (o *OnyxLogEmbed) SetFooter(url string) {
	o.embed.SetFooter("Onyx", url)
	o.embed.SetTimestamp(o.time)
}
