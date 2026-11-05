package mcp

import (
	"encoding/json"
	"fmt"
)

// message is any JSON-RPC 2.0 message: a request (ID and Method), a
// notification (Method only) or a response (ID with Result or Error).
// One shape for all of them keeps decoding a single step.
type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error returned by an MCP server.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const (
	jsonRPCVersion = "2.0"
	// codeMethodNotFound answers server requests this client does not serve.
	codeMethodNotFound = -32601
)

// Error formats the error with its code, so logs show what the server meant.
func (e *RPCError) Error() string {
	return fmt.Sprintf("mcp error %d: %s", e.Code, e.Message)
}

// isResponse reports whether msg answers a request.
func (msg message) isResponse() bool {
	return len(msg.ID) > 0 && msg.Method == ""
}

// isRequest reports whether msg is a request the server sent us.
func (msg message) isRequest() bool {
	return len(msg.ID) > 0 && msg.Method != ""
}

// replyToServer builds the answer to a server-initiated request. The client
// declares no capabilities, so only ping is served.
func replyToServer(req message) message {
	reply := message{JSONRPC: jsonRPCVersion, ID: req.ID}
	if req.Method == "ping" {
		reply.Result = json.RawMessage(`{}`)
		return reply
	}
	reply.Error = &RPCError{Code: codeMethodNotFound, Message: "method not found: " + req.Method}
	return reply
}

// newMessage builds a request (when id is set) or a notification.
func newMessage(id json.RawMessage, method string, params any) (message, error) {
	msg := message{JSONRPC: jsonRPCVersion, ID: id, Method: method}
	if params == nil {
		return msg, nil
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return msg, fmt.Errorf("encode %s params: %w", method, err)
	}
	msg.Params = raw
	return msg, nil
}

// resultOf returns a response's result, or its error.
func resultOf(resp message) (json.RawMessage, error) {
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}
