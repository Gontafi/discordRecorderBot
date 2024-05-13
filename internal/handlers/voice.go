package handlers

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"log"
	"time"
)

const recordingTimeLimit = 10 * time.Minute

var BotUserID string

var (
	voiceConnections  = make(map[string]*discordgo.VoiceConnection)
	voiceStopChannels = make(map[string]chan struct{})
)

// createPionRTPPacket converts a Discordgo audio packet into a Pion RTP packet.
func createPionRTPPacket(p *discordgo.Packet) *rtp.Packet {
	return &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    0x78,
			SequenceNumber: p.Sequence,
			Timestamp:      p.Timestamp,
			SSRC:           p.SSRC,
		},
		Payload: p.Opus,
	}
}

// handleVoice handles incoming voice packets and writes them to .ogg files.
func handleVoice(c chan *discordgo.Packet, stopChan chan struct{}) {
	files := make(map[uint32]media.Writer)

	for {
		select {
		case p, ok := <-c:
			if !ok {
				return // Channel closed
			}

			file, ok := files[p.SSRC]
			if !ok {
				var err error
				file, err = oggwriter.New(fmt.Sprintf("%d-%d.ogg", p.SSRC, time.Now().Unix()), 48000, 2)
				if err != nil {
					fmt.Printf("Failed to create file for SSRC %d: %v\n", p.SSRC, err)
					continue
				}
				files[p.SSRC] = file
			}

			rtpPacket := createPionRTPPacket(p)
			if err := file.WriteRTP(rtpPacket); err != nil {
				fmt.Printf("Failed to write RTP to file for SSRC %d: %v\n", p.SSRC, err)
			}
		case <-stopChan:
			for _, f := range files {
				f.Close()
			}
			return
		}
	}
}

// OnVoiceStateUpdate handles voice state updates to join new voice channels and start recording.
func OnVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	if BotUserID == "" {
		u, err := s.User("@me")
		if err != nil {
			log.Printf("Cannot fetch bot user details: %v", err)
			return
		}
		BotUserID = u.ID
	}

	// Ignore bot's own voice states
	if vs.UserID == BotUserID {
		return
	}

	// Check if the bot is already connected to the voice channel
	if vc, ok := voiceConnections[vs.GuildID]; ok && vc.ChannelID == vs.ChannelID {
		return
	}

	guild, err := s.State.Guild(vs.GuildID)
	if err != nil {
		log.Printf("Error retrieving guild: %v", err)
		return
	}

	// Count non-bot users in the new channel
	nonBotUserCount := 0
	for _, voiceState := range guild.VoiceStates {
		if voiceState.ChannelID == vs.ChannelID {
			member, err := s.GuildMember(vs.GuildID, voiceState.UserID)
			if err != nil {
				log.Printf("Error retrieving member information: %v", err)
				continue
			}
			if !member.User.Bot {
				nonBotUserCount++
			}
		}
	}

	if nonBotUserCount > 0 {
		// Join the channel and start recording
		vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, false)
		if err != nil {
			log.Printf("Error joining voice channel: %v", err)
			return
		}
		voiceConnections[vs.GuildID] = vc
		log.Printf("Joined voice channel '%s' in guild '%s'", vs.ChannelID, vs.GuildID)

		// Start a goroutine to handle voice packets
		stopChan := make(chan struct{})
		go handleVoice(vc.OpusRecv, stopChan)

		// Wait for the predefined time limit or until the channel is empty
		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			defer ticker.Stop()

		Loop:
			for {
				select {
				case <-ticker.C:
					break Loop
				case <-time.After(1 * time.Second):
					// Refresh the guild object to get the latest voice states
					guild, err = s.State.Guild(vs.GuildID)
					if err != nil {
						log.Printf("Error retrieving guild for voice state check: %v", err)
						break Loop
					}

					// Check if the bot is the only one in the channel
					nonBotUserCount = 0
					for _, voiceState := range guild.VoiceStates {
						if voiceState.ChannelID == vc.ChannelID && voiceState.UserID != BotUserID {
							nonBotUserCount++
						}
					}

					if nonBotUserCount == 0 {
						break Loop
					}
				}
			}

			vc.Disconnect()
			close(vc.OpusRecv)
			delete(voiceConnections, vs.GuildID)
			close(stopChan)
			log.Printf("Stopped recording in guild '%s'", vs.GuildID)
		}()
	} else {
		if vc, ok := voiceConnections[vs.GuildID]; ok {
			vc.Disconnect()
			delete(voiceConnections, vs.GuildID)
			log.Printf("Left voice channel '%s' in guild '%s'", vs.ChannelID, vs.GuildID)
		}
	}
}
