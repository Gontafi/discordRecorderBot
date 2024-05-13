package handlers

import (
	"discord_recorder_bot/pkg"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
)

func init() { flag.Parse() }

func InitializeCommands(s *discordgo.Session, GuildID string) ([]*discordgo.ApplicationCommand, error) {

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(pkg.Commands))
	for i, v := range pkg.Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return registeredCommands, nil
}
