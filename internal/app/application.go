package app

import (
	"discord_recorder_bot/internal/discord"
	"discord_recorder_bot/internal/handlers"
	"discord_recorder_bot/pkg"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"sync"
)

func StartApplication() (*discordgo.Session, map[string][]*discordgo.ApplicationCommand, error) {
	log.Println("reading config...")
	err := pkg.ReadCredentials()
	if err != nil {
		return nil, nil, err
	}

	log.Println(pkg.Config.DiscordBotToken)
	if pkg.Config.DiscordBotToken == "" {
		return nil, nil, errors.New("No token provided")
	}
	s, err := discord.CreateSession(pkg.Config.DiscordBotToken)
	if err != nil {
		return nil, nil, err
	}

	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates | discordgo.IntentsGuilds)

	s.AddHandler(handlers.OnVoiceStateUpdate)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := pkg.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	log.Println("authentication bot...")
	err = s.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return nil, nil, err
	}

	handlers.BotUserID = s.State.User.ID
	//guildIDs := discord.GetGuildIDs(s)
	//
	//var (
	//	wg          sync.WaitGroup
	//	mu          sync.Mutex
	//	commandsMap = make(map[string][]*discordgo.ApplicationCommand)
	//	firstError  error
	//)
	//
	//log.Println("adding commands...")
	//for _, guildID := range guildIDs {
	//	wg.Add(1)
	//	go func(gID string) {
	//		defer wg.Done()
	//
	//		registeredCommands, err := handlers.InitializeCommands(s, gID)
	//		if err != nil {
	//			mu.Lock()
	//			if firstError == nil { // Store the first error that occurs
	//				firstError = err
	//			}
	//			mu.Unlock()
	//			return
	//		}
	//
	//		mu.Lock()
	//		commandsMap[gID] = registeredCommands
	//		mu.Unlock()
	//	}(guildID)
	//}
	//if firstError != nil {
	//	return nil, nil, firstError
	//}

	//wg.Wait() // Wait for all goroutines to finish

	log.Println("bot started successfully")
	return s, nil, nil
}

func StopApplication(s *discordgo.Session, commandsMap map[string][]*discordgo.ApplicationCommand) error {
	log.Println("Removing commands...")

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		firstError error
	)
	guildIDs := discord.GetGuildIDs(s)

	for _, guildID := range guildIDs {
		wg.Add(1)
		go func(gID string) {
			defer wg.Done()

			for _, v := range commandsMap[gID] {
				err := s.ApplicationCommandDelete(s.State.User.ID, gID, v.ID)
				if err != nil {
					log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
				}
				if err != nil {
					mu.Lock()
					if firstError == nil { // Store the first error that occurs
						firstError = err
					}
					mu.Unlock()
					return
				}
			}

		}(guildID)
	}

	wg.Wait()

	if firstError != nil {
		return firstError
	}

	return nil
}
