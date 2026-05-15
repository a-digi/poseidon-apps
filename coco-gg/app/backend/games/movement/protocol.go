package movement

type MessageType string

const (
	MsgHello    MessageType = "hello"
	MsgInput    MessageType = "input"
	MsgWelcome  MessageType = "welcome"
	MsgSnapshot MessageType = "snapshot"
	MsgLeft     MessageType = "left"
	MsgError    MessageType = "error"
)

// Hello is the first message a client must send. Room is REQUIRED and must be
// non-empty: it names an existing room to join. Rooms can only be created via
// the HTTP POST /api/rooms endpoint; an empty or unknown Room causes the server
// to reply with an error and close. Name is required: after trimming whitespace
// it must be 1-32 characters.
type Hello struct {
	Type MessageType `json:"type"`
	Room string      `json:"room"`
	Name string      `json:"name"`
}

type Input struct {
	Type  MessageType `json:"type"`
	Up    bool        `json:"up"`
	Down  bool        `json:"down"`
	Left  bool        `json:"left"`
	Right bool        `json:"right"`
	Seq   int         `json:"seq"`
}

type Bounds struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type Welcome struct {
	Type     MessageType `json:"type"`
	PlayerID string      `json:"playerId"`
	Room     string      `json:"room"`
	TickHz   int         `json:"tickHz"`
	Arena    Bounds      `json:"arena"`
}

type PlayerSnapshot struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
}

type Snapshot struct {
	Type    MessageType      `json:"type"`
	Tick    int              `json:"tick"`
	Players []PlayerSnapshot `json:"players"`
}

type Left struct {
	Type     MessageType `json:"type"`
	PlayerID string      `json:"playerId"`
}

type ErrorMsg struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}
