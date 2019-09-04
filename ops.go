package discord

type GatewayEventType string

// Gateway operations
const (
	GatewayHello                    GatewayEventType = "HELLO"
	GatewayReady                    GatewayEventType = "READY"
	GatewayResume                   GatewayEventType = "RESUME"
	GatewayChannelCreate            GatewayEventType = "CHANNEL_CREATE"
	GatewayChannelUpdate            GatewayEventType = "CHANNEL_UPDATE"
	GatewayChannelDelete            GatewayEventType = "CHANNEL_DELETE"
	GatewayChannelPinsUpdate        GatewayEventType = "CHANNEL_PINS_UPDATE"
	GatewayGuildCreate              GatewayEventType = "GUILD_CREATE"
	GatewayGuildUpdate              GatewayEventType = "GUILD_UPDATE"
	GatewayGuildDelete              GatewayEventType = "GUILD_DELETE"
	GatewayGuildBanAdd              GatewayEventType = "GUILD_BAN_ADD"
	GatewayGuildBanRemove           GatewayEventType = "GUILD_BAN_REMOVE"
	GatewayGuildEmojisUpdate        GatewayEventType = "GUILD_EMOJIS_UPDATE"
	GatewayGuildIntegrationsUpdate  GatewayEventType = "GUILD_INTEGRATIONS_UPDATE"
	GatewayGuildMemberAdd           GatewayEventType = "GUILD_MEMBER_ADD"
	GatewayGuildMemberRemove        GatewayEventType = "GUILD_MEMBER_REMOVE"
	GatewayGuildMemberUpdate        GatewayEventType = "GUILD_MEMBER_UPDATE"
	GatewayGuildMembersChunk        GatewayEventType = "GUILD_MEMBERS_CHUNK"
	GatewayGuildRoleCreate          GatewayEventType = "GUILD_ROLE_CREATE"
	GatewayGuildRoleUpdate          GatewayEventType = "GUILD_ROLE_UPDATE"
	GatewayGuildRoleDelete          GatewayEventType = "GUILD_ROLE_DELETE"
	GatewayMessageCreate            GatewayEventType = "MESSAGE_CREATE"
	GatewayMessageUpdate            GatewayEventType = "MESSAGE_UPDATE"
	GatewayMessageDelete            GatewayEventType = "MESSAGE_DELETE"
	GatewayMessageDeleteBulk        GatewayEventType = "MESSAGE_DELETE_BULK"
	GatewayMessageReactionAdd       GatewayEventType = "MESSAGE_REACTION_ADD"
	GatewayMessageReactionRemove    GatewayEventType = "MESSAGE_REACTION_REMOVE"
	GatewayMessageReactionRemoveAll GatewayEventType = "MESSAGE_REACTION_REMOVE_ALL"
	GatewayPresenceUpdate           GatewayEventType = "PRESENCE_UPDATE"
	GatewayTypingStart              GatewayEventType = "TYPING_START"
	GatewayUserUpdate               GatewayEventType = "USER_UPDATE"
	GatewayVoiceStateUpdate         GatewayEventType = "VOICE_STATE_UPDATE"
	GatewayVoiceServerUpdate        GatewayEventType = "VOICE_SERVER_UPDATE"
	GatewayWebhooksUpdate           GatewayEventType = "WEBHOOKS_UPDATE"
)

type gatewayOp struct {
	Op int              `json:"op"`
	D  interface{}      `json:"d"`
	S  int              `json:"s"`
	T  GatewayEventType `json:"t"`
}

type helloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type identifyData struct {
	Token      string `json:"token"`
	Properties struct {
		OS      string `json:"$os"`
		Browser string `json:"$browser"`
		Device  string `json:"$device"`
	} `json:"properties"`
}

type resumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int    `json:"seq"`
}

type dispatchData struct {
	D struct {
		SessionID string `json:"session_id"`
		User      User   `json:"user"`
	} `json:"d"`
}

type GuildMemberUpdate struct {
	GuildID string `json:"guild_id"`
	Roles   []Role `json:"roles"`
	User    User   `json:"user"`
	Nick    string `json:"nick"`
}
