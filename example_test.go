package discord_test

import (
	"fmt"
	"github.com/miniriley2012/discord"
)

func ExampleClient_Listen() {
	client := discord.NewClient("TOKEN")

	if err := client.Connect(); err != nil {
		panic(err)
	}

	// Add handlers...

	if err := client.Listen(); err != nil {
		panic(err)
	}
}

func ExampleClient_Handle() {
	client.Handle(discord.GatewayMessageCreate, discord.MessageHandler(func(client *discord.Client, message discord.Message) {
		fmt.Printf("%v: %v\n", message.Author.Username, message.Content)
	}))
}

func Example() {
	client := discord.NewClient("TOKEN")

	if err := client.Connect(); err != nil {
		panic(err)
	}

	client.Handle(discord.GatewayMessageCreate, discord.MessageHandler(func(client *discord.Client, message discord.Message) {
		fmt.Printf("%v: %v\n", message.Author.Username, message.Content)
	}))

	if err := client.Listen(); err != nil {
		panic(err)
	}
}
