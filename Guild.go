package discord

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// TODO add Discord constants

// Guild is the Go representation of Guild in the Discord API
type Guild struct {
	ID                          string
	Name                        string
	Icon                        string
	Owner                       bool
	OwnerID                     string `mapstructure:"owner_id"`
	Permissions                 int
	Region                      string
	AFKChannelID                string `mapstructure:"afk_channel_id"`
	AFKTimeout                  int    `mapstructure:"afk_timeout"`
	EmbedEnabled                bool   `mapstructure:"embed_enabled"`
	EmbedChannelID              string `mapstructure:"embed_channel_id"`
	VerificationLevel           int    `mapstructure:"verification_level"`
	DefaultMessageNotifications int    `mapstructure:"default_message_notifications"`
	ExplicitContentFilter       int    `mapstructure:"explicit_content_filter"`
	Roles                       []Role
	Emojis                      []Emoji
	Features                    []string
	MFALevel                    int    `mapstructure:"mfa_level"`
	ApplicationID               string `mapstructure:"application_id"`
	WidgetEnabled               bool   `mapstructure:"widget_enabled"`
	WidgetChannelID             string `mapstructure:"widget_channel_id"`
	SystemChannelID             string `mapstructure:"system_channel_id"`
	MaxPresences                int    `mapstructure:"max_presences"`
	MaxMembers                  int    `mapstructure:"max_members"`
	VanityURLCode               string `mapstructure:"vanity_url_code"`
	Description                 string
	Banner                      string
	PremiumTier                 int    `mapstructure:"premium_tier"`
	PremiumSubscriptionCount    int    `mapstructure:"premium_subscription_count"`
	PreferredLocale             string `mapstructure:"preferred_locale"`
	client                      *Client
}

type GuildCreate struct {
	Guild
	JoinedAt    time.Time `mapstructure:"joined_at"`
	Large       bool
	Unavailable bool
	MemberCount int          `mapstructure:"member_count"`
	VoiceStates []VoiceState `mapstructure:"voice_states"`
	Members     []GuildMember
	Channels    []Channel
	Presences   []Presence `mapstructure:"presences"`
}

func (g *Guild) Members(limit int) ([]GuildMember, error) {
	url := "https://discordapp.com/api/v6/guilds/" + g.ID + "/members"
	if limit > 1 {
		url += "?limit=" + strconv.Itoa(limit)
	}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+g.client.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var members []GuildMember

	if err = json.Unmarshal(b, &members); err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.GuildID == "" {
			m.GuildID = g.ID
		}
	}

	return members, nil
}

func (g *Guild) Member(id string) (member GuildMember, err error) {
	req, err := http.NewRequest(http.MethodGet, "https://discordapp.com/api/v6/guilds/"+g.ID+"/members/"+id, http.NoBody)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bot "+g.client.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &member); err != nil {
		return
	}

	if member.GuildID == "" {
		member.GuildID = g.ID
	}

	return
}

// TODO Finish VoiceState
type VoiceState struct{}

// Stores channels
type GuildStore map[string]Guild

// Get a channel by ID
func (store GuildStore) Get(id string) *Guild {
	if g, ok := store[id]; ok {
		return &g
	}
	return nil
}

// Add a channel to the store
func (store GuildStore) Add(guild Guild) {
	store[guild.ID] = guild
}
