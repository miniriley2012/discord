package discord

import (
	"github.com/mitchellh/mapstructure"
)

type EventHandler interface {
	Handle(*Client, map[string]interface{}) error
}

type MessageHandler func(client *Client, message Message)

func (h MessageHandler) Handle(client *Client, i map[string]interface{}) error {
	var m Message
	if err := mapstructure.Decode(i, &m); err != nil {
		return err
	}
	h(client, m)
	return nil
}

type PresenceHandler func(client *Client, presence Presence)

func (ph PresenceHandler) Handle(client *Client, data map[string]interface{}) error {
	var p Presence
	if err := mapstructure.Decode(data, &p); err != nil {
		return err
	}
	ph(client, p)
	return nil
}

type UserUpdateHandler func(client *Client, user User)

func (handler UserUpdateHandler) Handle(client *Client, data map[string]interface{}) error {
	var user User
	if err := mapstructure.Decode(data, &user); err != nil {
		return err
	}
	handler(client, user)
	return nil
}

type GuildMemberUpdateHandler func(client *Client, update GuildMemberUpdate)

func (h GuildMemberUpdateHandler) Handle(client *Client, data map[string]interface{}) error {
	var update GuildMemberUpdate
	if err := mapstructure.Decode(data, &update); err != nil {
		return err
	}
	h(client, update)
	return nil
}
