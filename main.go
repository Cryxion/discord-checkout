package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	botToken := os.Getenv("BOT_TOKEN")

	dg, err := discordgo.New(botToken)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	dg.AddHandler(onInteractionCreate)
	dg.AddHandler(onMessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	// wait for a termination signal, stay active until interrupted
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	println("Message received:", m.Content)
	if m.Author.Bot {
		return
	}

	// check message author for role PaymentManager
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		fmt.Println("Error fetching member:", err)
		return
	}
	// Check if the member has the "PaymentManager" role
	hasRole := false
	roles, err := s.GuildRoles(m.GuildID)
	for _, roleID := range member.Roles {
		if err != nil {
			return
		}
		for _, role := range roles {
			if role.ID == roleID {
				if role.Name == "PaymentManager" {
					hasRole = true
					break
				}
			}
		}
	}
	if !hasRole {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Create payment button if the message starts with "!sp"
	if strings.HasPrefix(m.Content, "!sp") {
		// get parameters from the message
		eventid := 0
		_, err := fmt.Sscanf(m.Content, "!sp %d", &eventid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid start payment command.")
			return
		}
		// might want to check if the eventid is valid here and get event name
		if eventid <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid event ID.")
			return
		}
		// For demonstration, we assume eventid is valid and proceed
		eventName := "Sample Event" // Replace with actual event name retrieval logic
		// Logic to check event ID
		fmt.Printf("Event ID: %d\n", eventid)

		componentsContinue := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Continue Payment for " + eventName,
						Style:    discordgo.PrimaryButton,
						CustomID: "continue_payment_" + strconv.Itoa(eventid),
					},
				},
			},
		}
		// Send a message
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    "Click the button below to complete your transaction:",
			Components: componentsContinue,
		})
		if err != nil {
			fmt.Println("Error sending message:", err)
		} else {
			fmt.Println("Message sent successfully!")
		}
	}
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "continue_payment_") {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please check your DMs to complete the payment.",
			},
		})

		dmUser(s, i.Member.User.ID, i.MessageComponentData().CustomID[len("continue_payment_"):])
	}
}

// so the bot won't reply or expose other user's discord IDs within the channel
func dmUser(s *discordgo.Session, userID string, skuID string) error {
	// Create a DM channel with the user
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	// ensure the skuID is valid, not empty
	if skuID == "" {
		return fmt.Errorf("invalid SKU ID provided")
	}

	// Send the message to the DM channel
	url := fmt.Sprintf("%s?discord_id=%s&sku_id=%s", os.Getenv("CHECKOUT_URL"), userID, skuID)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Complete Payment",
					Style: discordgo.LinkButton,
					URL:   url,
				},
			},
		},
	}

	// let user know that the bot is sending a DM
	_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content:    "Click the button below to complete your transaction:",
		Components: components,
	})

	if err != nil {
		return fmt.Errorf("failed to send DM: %w", err)
	}

	return nil
}
