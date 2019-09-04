package discord_test

import (
	"fmt"
	"github.com/miniriley2012/discord"
	"os"
	"testing"
)

var client *discord.Client

func TestConnection(t *testing.T) {
	token := os.Getenv("TOKEN")
	client = discord.NewClient(token)
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
}

func TestListen(t *testing.T) {
	client.Handle(discord.GatewayMessageCreate, discord.MessageHandler(func(client *discord.Client, msg discord.Message) {
		fmt.Printf("%v: %v\n", msg.Author.Username, msg.Content)
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}
	}))

	fmt.Println("Listening for message...")
	if err := client.Listen(); err != nil {
		t.Fatal(err)
	}
}
