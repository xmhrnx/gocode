package wire

// ChatList is the desktop's open chats.
type ChatList struct {
	Desktop string        `json:"desktop"`
	Chats   []ChatSummary `json:"chats"`
}

// ChatSummary is one chat in the list.
type ChatSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Folder string `json:"folder"`
	Busy   bool   `json:"busy"`
	// Waiting is true while the agent waits for a permission answer.
	Waiting bool `json:"waiting"`
}

// ChatView is what a phone shows of a chat: its latest rows and, while
// it works, the run's progress.
type ChatView struct {
	ChatSummary
	// Rows are the latest rows of the transcript, oldest first; Earlier
	// counts the rows left out before them.
	Rows    []Row `json:"rows"`
	Earlier int   `json:"earlier"`
	// Status describes the run: "Working", a retry, or "" when idle.
	Status string `json:"status,omitempty"`
	// Seconds is how long the run has worked, and TokensPerSecond how
	// fast its model writes.
	Seconds         int     `json:"seconds,omitempty"`
	TokensPerSecond float64 `json:"tokens_per_second,omitempty"`
	// Permission is the request waiting for an answer, if any.
	Permission *Permission `json:"permission,omitempty"`
	// Queued counts prompts waiting for the agent's next step.
	Queued int `json:"queued,omitempty"`
}

// Row kinds.
const (
	RowUser      = "user"
	RowText      = "text"
	RowReasoning = "reasoning"
	RowTool      = "tool"
	RowNotice    = "notice"
	RowMedia     = "media"
)

// Row is one row of a transcript.
type Row struct {
	Kind string `json:"kind"`
	// Text is the message, reasoning or notice; for a tool, what it does
	// ("Read main.go"); for media, its file name.
	Text string `json:"text"`
	// Tool rows: the tool's name and its state: "running", "done" or
	// "failed".
	Tool  string `json:"tool,omitempty"`
	State string `json:"state,omitempty"`
}

// Tool states.
const (
	StateRunning = "running"
	StateDone    = "done"
	StateFailed  = "failed"
)

// Permission is a tool call waiting for the user's approval.
type Permission struct {
	Tool    string `json:"tool"`
	Summary string `json:"summary"`
	// Detail is the command, the diff or the arguments.
	Detail string `json:"detail,omitempty"`
}
