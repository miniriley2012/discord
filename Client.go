package discord

import (
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
	Nickname     string    `json:"nick"`
	Roles        []Role    `json:"roles"`
	JoinedAt     time.Time `json:"joined_at"`
	PremiumSince time.Time `json:"premium_since"`
	Deaf         bool      `json:"deaf"`
	Mute         bool      `json:"mute"`
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
}

// Creates a new Discord Client.
func NewClient(token string) *Client {
	return &Client{
		Token:        token,
		handlers:     map[GatewayEventType]EventHandler{},
		channelStore: ChannelStore{},
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
