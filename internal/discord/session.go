package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func CreateSession(token string) (*discordgo.Session, error) {
	s, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		fmt.Println("error creating Discord session:", err)
		return nil, err
	}

	return s, nil
}

func GetGuildIDs(s *discordgo.Session) []string {
	var guildIDs []string
	for _, guild := range s.State.Guilds {
		guildIDs = append(guildIDs, guild.ID)
	}
	return guildIDs
}
