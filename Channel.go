package discord

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ChannelType int

var rateLimiter chan struct{}
var dummy struct{}

// Channel types
const (
	ChannelTypeGuildText = ChannelType(iota)
	ChannelTypeDM
	ChannelTypeGuildVoice
	ChannelTypeGroupDM
	ChannelTypeGuildCategory
	ChannelTypeGuildNews
	ChannelTypeGuildStore
)

func init() {
	rateLimiter = make(chan struct{}, 5)
	go func() {
		for {
			select {
			case <-rateLimiter:
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

// Channel is the Go representation of Channel in Discord's API.
type Channel struct {
	ID                   string
	Type                 ChannelType
	GuildID              string
	Position             int
	PermissionOverwrites []Overwrite
	Name                 string
	Topic                string
	NSFW                 bool
	LastMessageID        string
	Bitrate              int
	UserLimit            int
	RateLimitPerUser     int
	Recipients           []User
	Icon                 string
	OwnerID              string
	ApplicationID        string
	ParentID             string
	LastPinTimestamp     time.Time
	client               *Client
}

// Sends a message to the channel
func (c *Channel) Send(message string) error {
	if message == "" {
		return errors.New("cannot send empty message")
	}

	req, err := http.NewRequest(http.MethodPost, "https://discordapp.com/api/v6/channels/"+c.ID+"/messages", strings.NewReader(`{"content":`+strconv.Quote(message)+`}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+c.client.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	rateLimiter <- dummy

	return resp.Body.Close()
}

// Messages returns the last limit messages in the channel
func (c *Channel) Messages(limit int) ([]Message, error) {
	url := "https://discordapp.com/api/v6/channels/" + c.ID + "/messages"
	if limit > 1 {
		url += "?limit=" + strconv.Itoa(limit)
	}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+c.client.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var messages []Message

	if err = json.Unmarshal(b, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Stores channels
type ChannelStore map[string]Channel

// Get a channel by ID
func (store ChannelStore) Get(id string) *Channel {
	if c, ok := store[id]; ok {
		return &c
	}
	return nil
}

// Add a channel to the store
func (store ChannelStore) Add(channel Channel) {
	store[channel.ID] = channel
}

// Permission Overwrite in Message
type Overwrite struct {
	ID    string
	Type  string
	Allow int
	Deny  int
}

// Message is the Go representation of Message in Discord's API.
type Message struct {
	ID              string
	ChannelID       string `mapstructure:"channel_id"`
	GuildID         string `mapstructure:"guild_id"`
	Author          User
	Content         string
	TTS             bool
	MentionEveryone bool `mapstructure:"mention_everyone"`
	Mentions        []User
	MentionRoles    []Role `mapstructure:"mention_roles"`
	Attachments     []Attachment
	Embeds          []Embed
	Reactions       []Reaction
	Nonce           string
	Pinned          bool
	WebhookID       string `mapstructure:"webhook_id"`
	Type            int // TODO Add MessageType type
	Flags           int // Add MessageFlag type
}

// Attachment is the Go representation of Attachment in Discord's API.
type Attachment struct {
	ID       string
	Filename string
	Size     int
	URL      string
	ProxyURL string
	Height   *int
	Width    *int
}

// Embed is the Go representation of Embed in Discord's API.
type Embed struct {
	Title       *string
	Type        *string
	Description *string
	URL         *string
	Timestamp   *time.Time
	Color       *int
}

// Emoji is the Go representation of Emoji in Discord's API.
type Emoji struct {
	ID            string
	Name          string
	Roles         []Role
	User          User
	RequireColons bool
	Managed       bool
	Animated      bool
}

// Reaction is the Go representation of Reaction in Discord's API.
type Reaction struct {
	Count int
	Me    bool
	Emoji Emoji
}
