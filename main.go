package main

import (
	"discord_recorder_bot/internal/app"
	"log"
	"os"
	"os/signal"
)

// initialize program and start listening a list of guilds for each channels
func main() {
	s, registeredCommands, err := app.StartApplication()
	if err != nil {
		log.Println(err)
		return
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	err = app.StopApplication(s, registeredCommands)
	if err != nil {
		log.Println(err)
	}

	log.Println("Gracefully shutting down.")
}
