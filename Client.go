package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"go/build"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// User is the Go representation of User in Discord's API.
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot"`
	PremiumType   int    `json:"premium_type"`
}

type RolePerm int64

// Role Permissions
const (
	PermCreateInstantInvite = 0x00000001
	PermKickMembers         = 0x2
	PermBanMembers          = 0x4
	PermAdministrator       = 0x8
	PermManageChannels      = 0x10
	PermManageGuild         = 0x20
	PermAddReactions        = 0x40
	PermViewAuditLog        = 0x80
	PermViewChannel         = 0x400
	PermSendMessages        = 0x800
	PermSendTTSMessages     = 0x1000
	PermManageMessages      = 0x2000
	PermEmbedLinks          = 0x4000
	PermAttachFiles         = 0x8000
	PermReadMessageHistory  = 0x10000
	PermMentionEveryone     = 0x20000
	PermUseExternalEmojis   = 0x40000
	PermConnect             = 0x100000
	PermSpeak               = 0x100000
	PermMuteMembers         = 0x400000
	PermDeafenMembers       = 0x800000
	PermMoveMembers         = 0x1000000
	PermUseVAD              = 0x2000000
	PermPrioritySpeaker     = 0x100
	PermStream              = 0x200
	PermChangeNickname      = 0x4000000
	PermManageNicknames     = 0x8000000
	PermManageRoles         = 0x10000000
	PermManageWebhooks      = 0x20000000
	PermManageEmojis        = 0x40000000
)

// Role is the Go representation of Role in Discord's API.
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Hoist       bool   `json:"hoist"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
	Managed     bool   `json:"managed"`
	Mentionable bool   `json:"mentionable"`
}

type Roles []Role

func NewRole(id string) Role {
	if v, ok := roles[id]; ok {
		return v
	}
	return Role{ID: id}
}

func (r *Roles) UnmarshalJSON(data []byte) error {
	var IDs []string
	if err := json.Unmarshal(data, &IDs); err != nil {
		return err
	}

	for _, id := range IDs {
		*r = append(*r, NewRole(id))
	}
	return nil
}

type ActivityType int

// Activity types
const (
	ActivityGame = iota
	ActivityStreaming
	ActivityListening
)

type ActivityFlag int

// Activity flags
const (
	ActivityInstanceFlag = 1 << iota
	ActivityJoinFlag
	ActivitySpectateFlag
	ActivityJoinRequestFlag
	ActivitySyncFlag
	ActivityPlayFlag
)

// ActivityTimestamps is the Go representation of ActivityTimestamps in the Discord API.
type ActivityTimestamps struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// ActivityParty is the Go representation of ActivityParty in the Discord API.
type ActivityParty struct {
	ID   string `json:"id"`
	Size []int  `json:"size"`
}

// ActivityAsset is the Go representation of ActivityAsset in the Discord API.
type ActivityAsset struct {
	LargeImage string `json:"large_image"`
	LargeText  string `json:"large_text"`
	SmallImage string `json:"small_image"`
	SmallText  string `json:"small_text"`
}

// ActivitySecret is the Go representation of ActivitySecret in the Discord API.
type ActivitySecret struct {
	Join     string `json:"join"`
	Spectate string `json:"spectate"`
	Match    string `json:"match"`
}

// Activity is the Go representation of Activity in the Discord API.
type Activity struct {
	Name          string             `json:"name"`
	Type          ActivityType       `json:"type"`
	URL           string             `json:"url"`
	Timestamps    ActivityTimestamps `json:"timestamps"`
	ApplicationID string             `json:"application_id"`
	Details       string             `json:"details"`
	State         string             `json:"state"`
	Party         ActivityParty      `json:"party"`
	Assets        []ActivityAsset    `json:"assets"`
	Secrets       []ActivitySecret   `json:"secrets"`
	Instance      bool               `json:"instance"`
	Flags         ActivityFlag       `json:"flags"`
}

// Presence is the Go representation of Presence in the Discord API.
type Presence struct {
	User         User       `json:"user"`
	Roles        []string   `json:"roles"`
	Game         Activity   `json:"game"`
	GuildID      string     `json:"guild_id"`
	Status       string     `json:"status"`
	Activities   []Activity `json:"activities"`
	ClientStatus struct {
		Desktop string `json:"desktop"`
		Mobile  string `json:"mobile"`
		Web     string `json:"web"`
	} `json:"client_status"`
}

// GuildMember is the Go representation of GuildMember in the Discord API.
type GuildMember struct {
	User         `json:"user"`
	GuildID      string    `json:"guild_id" mapstructure:"guild_id"`
	Nickname     string    `json:"nick"`
	Roles        Roles     `json:"roles"`
	JoinedAt     time.Time `json:"joined_at"`
	PremiumSince time.Time `json:"premium_since"`
	Deaf         bool      `json:"deaf"`
	Mute         bool      `json:"mute"`
}

func (g GuildMember) AddRole(client *Client, id string) error {
	var role *Role

	guild, err := client.GetGuild(g.GuildID)
	if err != nil {
		return err
	}

	for _, r := range guild.Roles {
		if r.ID == id {
			role = &r
			break
		}
	}

	if role == nil {
		return errors.New("role not found on this server")
	}

	g.Roles = append(g.Roles, *role)

	var roles []string
	for _, r := range g.Roles {
		roles = append(roles, r.ID)
	}

	b, _ := json.Marshal(struct {
		Roles []string `json:"roles"`
	}{roles})

	req, err := http.NewRequest(http.MethodPatch, "https://discordapp.com/api/v6/guilds/"+g.GuildID+"/members/"+g.ID, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+client.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

func (g *GuildMember) HasPermission(perm RolePerm) bool {
	for _, r := range g.Roles {
		if RolePerm(r.Permissions)&perm == perm {
			return true
		}
	}
	return false
}

// Client interacts with the Discord API.
type Client struct {
	User
	ws                *websocket.Conn
	heartbeatInterval int
	sequence          int
	sessionID         string
	Token             string
	handlers          map[GatewayEventType]EventHandler
	channelStore      ChannelStore
	guildStore        GuildStore
}

// Creates a new Discord Client.
func NewClient(token string) *Client {
	return &Client{
		Token:        token,
		handlers:     map[GatewayEventType]EventHandler{},
		channelStore: ChannelStore{},
		guildStore:   GuildStore{},
	}
}

//func setField(data interface{}, name string, value interface{}) error {
//	structValue := reflect.ValueOf(data).Elem()
//	structFieldValue := structValue.FieldByName(name)
//
//	if !structFieldValue.IsValid() {
//		return fmt.Errorf("no such field: %v", name)
//	}
//
//	if !structFieldValue.CanSet() {
//		return fmt.Errorf("cannot set field: %v", name)
//	}
//
//	structFieldType := structFieldValue.Type()
//	val := reflect.ValueOf(value)
//	if structFieldType != val.Type() {
//		return errors.New("value type does not match field type")
//	}
//
//	structFieldValue.Set(val)
//	return nil
//}

// Send heartbeats forever
func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Duration(c.heartbeatInterval) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			packet := gatewayOp{Op: 1}
			if c.sequence != 0 {
				packet.D = c.sequence
			}
			_ = c.ws.WriteJSON(&packet)
		}
	}
}

// Identify with the discord Gateway
func (c *Client) identify() error {
	data := identifyData{
		Token: c.Token,
		Properties: struct {
			OS      string `json:"$os"`
			Browser string `json:"$browser"`
			Device  string `json:"$device"`
		}{
			OS:      build.Default.GOOS,
			Browser: "discord-go",
			Device:  "discord-go",
		},
	}

	packet := gatewayOp{
		Op: 2,
		D:  data,
	}

	return c.ws.WriteJSON(packet)
}

// Resume session if disconnected
func (c *Client) resume() error {
	data := resumeData{
		Token:     c.Token,
		SessionID: c.sessionID,
		Sequence:  c.sequence,
	}

	packet := gatewayOp{
		Op: 6,
		D:  data,
	}

	return c.ws.WriteJSON(packet)
}

// Connect connects the client to a Discord Gateway.
func (c *Client) Connect() (err error) {
	c.ws, _, err = websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=6&encoding=json", nil)
	if err != nil {
		return err
	}

	var op gatewayOp
	if err := c.ws.ReadJSON(&op); err != nil {
		return err
	}

	c.heartbeatInterval = int(op.D.(map[string]interface{})["heartbeat_interval"].(float64))

	go c.heartbeat()

	if c.sequence == 0 {
		err = c.identify()
	} else {
		err = c.resume()
	}
	if err != nil {
		return err
	}

	var data dispatchData
	err = c.ws.ReadJSON(&data)
	if err != nil {
		return err
	}

	c.sessionID = data.D.SessionID
	c.User = data.D.User

	return nil
}

// handle chooses executes the handler for op
func handle(op GatewayEventType, c *Client, data map[string]interface{}) error {
	if v, ok := c.handlers[op]; ok {
		if err := v.Handle(c, data); err != nil {
			return err
		}
	}
	return nil
}

// Listen is a blocking function that will begin listening for Discord Gateway events.
func (c *Client) Listen() (err error) {
	for {
		var op gatewayOp
		if err = c.ws.ReadJSON(&op); err != nil {
			if e := err.(*net.OpError); e.Op == "read" {
				return nil
			}
			return err
		}

		if op.Op != 0 {
			continue
		}

		data, ok := op.D.(map[string]interface{})
		if !ok {
			continue
		}

		if err = handle(op.T, c, data); err != nil {
			return err
		}

		//switch op.T {
		//case GatewayChannelCreate:
		//case GatewayChannelUpdate:
		//case GatewayChannelDelete:
		//case GatewayChannelPinsUpdate:
		//case GatewayGuildCreate:
		//case GatewayGuildUpdate:
		//case GatewayGuildDelete:
		//case GatewayGuildBanAdd:
		//case GatewayGuildBanRemove:
		//case GatewayGuildEmojisUpdate:
		//case GatewayGuildIntegrationsUpdate:
		//case GatewayGuildMemberAdd:
		//case GatewayGuildMemberRemove:
		//case GatewayGuildMemberUpdate:
		//case GatewayGuildMembersChunk:
		//case GatewayGuildRoleCreate:
		//case GatewayGuildRoleUpdate:
		//case GatewayGuildRoleDelete:
		//case GatewayMessageCreate:
		//case GatewayMessageUpdate:
		//case GatewayMessageDelete:
		//case GatewayMessageDeleteBulk:
		//case GatewayMessageReactionAdd:
		//case GatewayMessageReactionRemove:
		//case GatewayMessageReactionRemoveAll:
		//case GatewayPresenceUpdate:
		//case GatewayTypingStart:
		//case GatewayUserUpdate:
		//case GatewayVoiceStateUpdate:
		//case GatewayVoiceServerUpdate:
		//case GatewayWebhooksUpdate:
		//}
	}
}

// Handle registers an EventHandler for eventType
func (c *Client) Handle(eventType GatewayEventType, handler EventHandler) {
	c.handlers[eventType] = handler
}

// Close closes the client's connection to the Discord Gateway.
func (c *Client) Close() error {
	if err := c.ws.WriteMessage(websocket.CloseMessage, nil); err != nil {
		return err
	}
	return c.ws.Close()
}

// readChannelResponse reads JSON into a Channel
func readChannelResponse(r io.Reader) (c Channel, err error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	var i map[string]interface{}

	if err = json.Unmarshal(b, &i); err != nil {
		return
	}

	if m, ok := i["message"]; ok {
		return c, errors.New(m.(string))
	}

	if id, ok := i["channel_id"].([]interface{}); ok {
		return c, errors.New(id[0].(string))
	}

	if err = json.Unmarshal(b, &c); err != nil {
		return
	}

	return
}

func readGuildResponse(r io.Reader) (g Guild, err error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	var i map[string]interface{}

	if err = json.Unmarshal(b, &i); err != nil {
		return
	}

	if m, ok := i["message"]; ok {
		return g, errors.New(m.(string))
	}

	if id, ok := i["guild_id"].([]interface{}); ok {
		return g, errors.New(id[0].(string))
	}

	if err = json.Unmarshal(b, &g); err != nil {
		return
	}

	return
}

// GetChannel returns a channel by ID
func (c *Client) GetChannel(id string) (ch Channel, err error) {
	channel := c.channelStore.Get(id)
	if channel == nil {
		var req *http.Request
		var resp *http.Response
		var channel Channel

		req, err = http.NewRequest(http.MethodGet, "https://discordapp.com/api/v6/channels/"+id, http.NoBody)
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bot "+c.Token)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		channel, err = readChannelResponse(resp.Body)
		if err != nil {
			return
		}
		c.channelStore.Add(channel)
		ch = channel
	} else {
		ch = *channel
	}
	ch.client = c
	return
}

// GetGuild returns a guild by ID
func (c *Client) GetGuild(id string) (g Guild, err error) {
	guild := c.guildStore.Get(id)
	if guild == nil {
		var req *http.Request
		var resp *http.Response
		var guild Guild

		req, err = http.NewRequest(http.MethodGet, "https://discordapp.com/api/v6/guilds/"+id, http.NoBody)
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bot "+c.Token)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		guild, err = readGuildResponse(resp.Body)
		if err != nil {
			return
		}

		for _, r := range guild.Roles {
			roles[r.ID] = r
		}

		c.guildStore.Add(guild)
		g = guild
	} else {
		g = *guild
	}
	g.client = c
	return
}
