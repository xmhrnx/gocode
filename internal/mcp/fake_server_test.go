package mcp

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"gocode/internal/settings"
)

// fakeServer is a scripted MCP server run by the test binary itself (the
// os/exec helper-process pattern), speaking stdio like a real one.
type fakeServer struct {
	in         *bufio.Scanner
	out        *json.Encoder
	toolsAdded bool
}

// fakeServerEnv makes the test binary act as the fake server.
const fakeServerEnv = "GOCODE_MCP_FAKE_SERVER"

func TestMain(m *testing.M) {
	if os.Getenv(fakeServerEnv) == "1" {
		runFakeServer()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeServerConfig configures the fake server under name.
func fakeServerConfig(name string) settings.MCPServer {
	return settings.MCPServer{Name: name, Transport: "stdio", Command: os.Args[0],
		// The race detector otherwise sleeps a second as the server exits.
		Env: map[string]string{fakeServerEnv: "1", "GORACE": "atexit_sleep_ms=0"}, Enabled: true}
}

func runFakeServer() {
	// Real servers sometimes print junk to stdout; the client must cope.
	fmt.Println("starting fake server")
	s := &fakeServer{in: bufio.NewScanner(os.Stdin), out: json.NewEncoder(os.Stdout)}
	for s.in.Scan() {
		var msg message
		if json.Unmarshal(s.in.Bytes(), &msg) == nil && msg.isRequest() {
			s.handle(msg)
		}
	}
}

func (s *fakeServer) handle(req message) {
	switch req.Method {
	case "initialize":
		s.reply(req, `{"protocolVersion":"2025-03-26","capabilities":{"tools":{"listChanged":true}},"serverInfo":{"name":"fake","version":"1"}}`)
	case "tools/list":
		s.listTools(req)
	case "tools/call":
		s.callTool(req)
	default:
		s.write(replyToServer(req))
	}
}

func (s *fakeServer) listTools(req message) {
	var params listToolsParams
	_ = json.Unmarshal(req.Params, &params)
	switch {
	case params.Cursor == "" && s.toolsAdded:
		s.reply(req, `{"tools":[{"name":"echo"},{"name":"added"}]}`)
	case params.Cursor == "":
		s.reply(req, `{"tools":[{"name":"echo","description":"Echoes text","inputSchema":{"type":"object","properties":{"text":{"type":"string"}}}}],"nextCursor":"page2"}`)
	default:
		s.reply(req, `{"tools":[{"name":"fail","annotations":{"readOnlyHint":true}}]}`)
	}
}

func (s *fakeServer) callTool(req message) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	_ = json.Unmarshal(req.Params, &params)
	switch params.Name {
	case "echo":
		image := base64.StdEncoding.EncodeToString(make([]byte, 2048))
		s.reply(req, fmt.Sprintf(`{"content":[{"type":"text","text":%q},{"type":"image","mimeType":"image/png","data":%q}]}`, params.Arguments["text"], image))
	case "fail":
		s.reply(req, `{"content":[{"type":"text","text":"boom"}],"isError":true}`)
	case "ask_client":
		ping, sampling := s.ask("ping"), s.ask("sampling/createMessage")
		s.reply(req, fmt.Sprintf(`{"content":[{"type":"text","text":"ping:%s sampling:%d"}]}`, ping.Result, sampling.Error.Code))
	case "change_tools":
		s.toolsAdded = true
		s.write(message{JSONRPC: jsonRPCVersion, Method: "notifications/tools/list_changed"})
		s.reply(req, `{"content":[{"type":"text","text":"changed"}]}`)
	case "crash":
		fmt.Fprintln(os.Stderr, "fatal: fake server crashed")
		os.Exit(3)
	case "slow":
		// Never answers, so the caller's deadline must end the call.
	}
}

// ask sends the client a request and reads its answer from stdin.
func (s *fakeServer) ask(method string) message {
	s.write(message{JSONRPC: jsonRPCVersion, ID: json.RawMessage(`"srv-` + method + `"`), Method: method})
	var resp message
	if s.in.Scan() {
		_ = json.Unmarshal(s.in.Bytes(), &resp)
	}
	if resp.Error == nil {
		resp.Error = &RPCError{}
	}
	return resp
}

func (s *fakeServer) reply(req message, result string) {
	s.write(message{JSONRPC: jsonRPCVersion, ID: req.ID, Result: json.RawMessage(strings.TrimSpace(result))})
}

func (s *fakeServer) write(msg message) {
	if err := s.out.Encode(msg); err != nil {
		os.Exit(2)
	}
}
