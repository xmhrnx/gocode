package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"
)

// answerTimeout bounds posting the reply to a server-initiated request.
const answerTimeout = 10 * time.Second

// errNoResponse reports a reply that ended before answering the request.
var errNoResponse = errors.New("MCP server sent no response to the request")

// readReply extracts the response to the request with id from an HTTP
// reply, which is either one JSON message or an SSE stream of them.
func (t *httpTransport) readReply(resp *http.Response, id json.RawMessage) (message, error) {
	if resp.StatusCode == http.StatusAccepted {
		return message{}, errNoResponse
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return message{}, fmt.Errorf("MCP server reply has no valid content type: %w", err)
	}
	switch mediaType {
	case "application/json":
		return decodeReply(resp.Body)
	case "text/event-stream":
		return t.readEventStream(resp.Body, id)
	default:
		return message{}, fmt.Errorf("MCP server replied with unexpected content type %q", mediaType)
	}
}

// decodeReply decodes a plain JSON reply.
func decodeReply(body io.Reader) (message, error) {
	var msg message
	if err := json.NewDecoder(body).Decode(&msg); err != nil {
		return message{}, fmt.Errorf("decode MCP server reply: %w", err)
	}
	return msg, nil
}

// readEventStream reads SSE events until the response to id arrives,
// routing the server's notifications and requests met on the way.
func (t *httpTransport) readEventStream(body io.Reader, id json.RawMessage) (message, error) {
	var found *message
	err := readSSE(body, func(data []byte) bool {
		var msg message
		if json.Unmarshal(data, &msg) != nil {
			return true // not a JSON-RPC message; keep reading
		}
		if msg.isResponse() && string(msg.ID) == string(id) {
			found = &msg
			return false
		}
		t.route(msg)
		return true
	})
	switch {
	case found != nil:
		return *found, nil
	case err != nil:
		return message{}, fmt.Errorf("read MCP event stream: %w", err)
	default:
		return message{}, errNoResponse
	}
}

// route handles a message from the stream that is not our response.
func (t *httpTransport) route(msg message) {
	switch {
	case msg.isRequest():
		go t.answer(msg)
	case msg.Method != "":
		t.events.notify(msg.Method, msg.Params)
	}
}

// answer posts the reply to a server-initiated request.
func (t *httpTransport) answer(req message) {
	ctx, cancel := context.WithTimeout(context.Background(), answerTimeout)
	defer cancel()
	// A failed reply leaves the server's request unanswered, which it
	// times out on its own; the client has nothing to recover.
	_ = t.send(ctx, replyToServer(req))
}
