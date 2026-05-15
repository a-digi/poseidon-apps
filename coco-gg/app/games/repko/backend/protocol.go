package repko

type MessageType string

const (
	MsgHello             MessageType = "hello"
	MsgPickCivilization  MessageType = "pick_civilization"
	MsgPickStartingTile  MessageType = "pick_starting_tile"
	MsgRecruit           MessageType = "recruit"
	MsgUpgrade           MessageType = "upgrade"
	MsgMove              MessageType = "move"
	MsgMarch             MessageType = "march"
	MsgAttack            MessageType = "attack"
	MsgBuyTile           MessageType = "buy_tile"
	MsgOfferDiplomacy    MessageType = "offer_diplomacy"
	MsgAcceptDiplomacy   MessageType = "accept_diplomacy"
	MsgDeclineDiplomacy  MessageType = "decline_diplomacy"
	MsgCancelDiplomacy   MessageType = "cancel_diplomacy"
	MsgUpgradeTile       MessageType = "upgrade_tile"
	MsgEndTurn           MessageType = "end_turn"
	MsgLeaveGame         MessageType = "leave_game"

	MsgWelcome MessageType = "welcome"
	MsgState   MessageType = "state"
	MsgError   MessageType = "error"
	MsgEvent   MessageType = "event"
)

type Hello struct {
	Type        MessageType `json:"type"`
	Room        string      `json:"room"`
	Name        string      `json:"name"`
	ResumeToken string      `json:"resumeToken,omitempty"`
}

type Welcome struct {
	Type        MessageType `json:"type"`
	PlayerID    string      `json:"playerId"`
	Room        string      `json:"room"`
	ResumeToken string      `json:"resumeToken"`
	You         WelcomeYou  `json:"you"`
}

type WelcomeYou struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ErrorMsg struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type State struct {
	Type             MessageType      `json:"type"`
	Phase            string           `json:"phase"`
	Board            *Board           `json:"board,omitempty"`
	Players          []*PlayerState   `json:"players"`
	Current          *CurrentTurn     `json:"currentTurn,omitempty"`
	Civilizations    []Civilization   `json:"civilizations,omitempty"`
	PendingDiplomacy []DiplomacyOffer `json:"pendingDiplomacy,omitempty"`
	Armies           []*Army          `json:"armies"`
	WinnerID         string           `json:"winnerId,omitempty"`
	MaxRounds        int              `json:"maxRounds,omitempty"`
	RoundNumber      int              `json:"roundNumber,omitempty"`
	You              *YouState        `json:"you,omitempty"`
}

type PickCivilization struct {
	Type           MessageType `json:"type"`
	CivilizationID string      `json:"civilizationId"`
}

type PickStartingTile struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type Recruit struct {
	Type  MessageType `json:"type"`
	Q     int         `json:"q"`
	R     int         `json:"r"`
	Unit  UnitType    `json:"unit"`
	Count int         `json:"count"`
}

type Upgrade struct {
	Type       MessageType `json:"type"`
	Q          int         `json:"q"`
	R          int         `json:"r"`
	StackIndex int         `json:"stackIndex"`
}

type Move struct {
	Type  MessageType `json:"type"`
	FromQ int         `json:"fromQ"`
	FromR int         `json:"fromR"`
	ToQ   int         `json:"toQ"`
	ToR   int         `json:"toR"`
	Units []StackPick `json:"units"`
}

type Attack struct {
	Type  MessageType `json:"type"`
	FromQ int         `json:"fromQ"`
	FromR int         `json:"fromR"`
	ToQ   int         `json:"toQ"`
	ToR   int         `json:"toR"`
	Units []StackPick `json:"units"`
}

type MarchMsg struct {
	Type  MessageType `json:"type"`
	FromQ int         `json:"fromQ"`
	FromR int         `json:"fromR"`
	ToQ   int         `json:"toQ"`
	ToR   int         `json:"toR"`
	Units []StackPick `json:"units"`
}

type BuyTile struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type UpgradeTileMsg struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type OfferDiplomacy struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type AcceptDiplomacy struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type DeclineDiplomacy struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type CancelDiplomacy struct {
	Type MessageType `json:"type"`
	Q    int         `json:"q"`
	R    int         `json:"r"`
}

type EndTurn struct {
	Type MessageType `json:"type"`
}

type LeaveGame struct {
	Type MessageType `json:"type"`
}

type GameEvent struct {
	Kind         string   `json:"kind"`
	ActorID      string   `json:"actorId"`
	ActorName    string   `json:"actorName"`
	TargetQ      *int     `json:"targetQ,omitempty"`
	TargetR      *int     `json:"targetR,omitempty"`
	FromQ        *int     `json:"fromQ,omitempty"`
	FromR        *int     `json:"fromR,omitempty"`
	DefenderID   string   `json:"defenderId,omitempty"`
	DefenderName string   `json:"defenderName,omitempty"`
	Unit         UnitType `json:"unit,omitempty"`
	UnitCount    int      `json:"unitCount,omitempty"`
	TileName     string   `json:"tileName,omitempty"`
	CivID        string   `json:"civId,omitempty"`
}

type EventEnvelope struct {
	Type  MessageType `json:"type"`
	Event GameEvent   `json:"event"`
}
