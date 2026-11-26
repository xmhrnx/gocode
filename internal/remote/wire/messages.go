// Package wire is the messages the phone and desktop exchange once their
// channel is open. Each is one JSON object with a type.
package wire

import "encoding/json"

// Envelope is any message: its type, the request it answers or makes,
// and the payload.
type Envelope struct {
	Type string `json:"type"`
	// ID matches a reply to its request; 0 for pushed updates.
	ID   int64           `json:"id,omitempty"`
	Body json.RawMessage `json:"body,omitempty"`
}

// Requests the phone sends.
const (
	// TypeList asks for the open chats; the reply is a ChatList, and
	// later changes to it are pushed while the phone watches nothing or
	// the list.
	TypeList = "list"
	// TypeWatch (Watch) asks for a chat's view; the reply is a ChatView,
	// and later changes are pushed until another watch or unwatch.
	TypeWatch   = "watch"
	TypeUnwatch = "unwatch"
	// TypeSend (Send) sends a prompt to a chat, queued if it is working.
	TypeSend = "send"
	// TypeStop (ChatRef) stops a chat's run.
	TypeStop = "stop"
	// TypeDecide (Decide) answers a chat's permission request.
	TypeDecide = "decide"
)

// Messages the desktop sends.
const (
	// TypeReply answers a request: Body is the result, or a Failure.
	TypeReply = "reply"
	// TypeChats (ChatList) and TypeChat (ChatView) push changes.
	TypeChats = "chats"
	TypeChat  = "chat"
)

// ChatRef names a chat.
type ChatRef struct {
	Chat string `json:"chat"`
}

// Watch asks to follow a chat.
type Watch = ChatRef

// Send is a prompt for a chat.
type Send struct {
	Chat string `json:"chat"`
	Text string `json:"text"`
}

// Decisions on a permission request.
const (
	DecisionAllow  = "allow"
	DecisionAlways = "always"
	DecisionDeny   = "deny"
)

// Decide answers a chat's permission request.
type Decide struct {
	Chat     string `json:"chat"`
	Decision string `json:"decision"`
	// Feedback tells the agent why, with a denial.
	Feedback string `json:"feedback,omitempty"`
}

// Failure is a reply to a request that couldn't be done.
type Failure struct {
	Error string `json:"error"`
}

// Hello is what a phone says in the handshake: its name, and the pairing
// token the first time it connects.
type Hello struct {
	Name  string `json:"name"`
	Token string `json:"token,omitempty"`
}

// Welcome is the desktop's answer to an accepted hello.
type Welcome struct {
	Name string `json:"name"`
}
