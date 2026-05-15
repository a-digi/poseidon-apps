package repko

type MessageType string

const (
	MsgHello             MessageType = "hello"
	MsgPickCivilization  MessageType = "pick_civilization"
	MsgPickStartingTile  MessageType = "pick_starting_tile"
	MsgRecruit           MessageType = "recruit"
	MsgUpgrade           MessageType = "upgrade"
	MsgMove              MessageType = "move"
	MsgAttack            MessageType = "attack"
	MsgBuyTile           MessageType = "buy_tile"
	MsgOfferDiplomacy    MessageType = "offer_diplomacy"
	MsgAcceptDiplomacy   MessageType = "accept_diplomacy"
	MsgDeclineDiplomacy  MessageType = "decline_diplomacy"
	MsgCancelDiplomacy   MessageType = "cancel_diplomacy"
	MsgEndTurn           MessageType = "end_turn"
	MsgLeaveGame         MessageType = "leave_game"

	MsgWelcome MessageType = "welcome"
	MsgState   MessageType = "state"
	MsgError   MessageType = "error"
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

type BuyTile struct {
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
