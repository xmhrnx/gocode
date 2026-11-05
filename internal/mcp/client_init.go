package mcp

import (
	"context"
	"fmt"
	"runtime/debug"
	"slices"
)

// initializeParams opens an MCP session. The client offers no
// capabilities: it neither exposes roots nor samples for the server.
type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    struct{}       `json:"capabilities"`
	ClientInfo      implementation `json:"clientInfo"`
}

// implementation names a client or server.
type implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// initializeResult is the part of the server's handshake answer we use.
type initializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
}

const (
	methodInitialize = "initialize"
	// latestProtocolVersion is the revision this client asks for.
	latestProtocolVersion = "2025-06-18"
	clientName            = "GoCode"
)

// supportedProtocolVersions lists the revisions a server may answer with.
var supportedProtocolVersions = []string{latestProtocolVersion, "2025-03-26", "2024-11-05"}

// initialize performs the handshake: agree on a protocol version, then
// tell the server the client is ready.
func (c *client) initialize(ctx context.Context) error {
	params := initializeParams{
		ProtocolVersion: latestProtocolVersion,
		ClientInfo:      implementation{Name: clientName, Version: clientVersion()},
	}
	var result initializeResult
	if err := c.request(ctx, methodInitialize, params, &result); err != nil {
		return err
	}
	if !slices.Contains(supportedProtocolVersions, result.ProtocolVersion) {
		return fmt.Errorf("MCP server speaks unsupported protocol version %q", result.ProtocolVersion)
	}
	c.transport.setProtocolVersion(result.ProtocolVersion)
	return c.notify(ctx, "notifications/initialized", nil)
}

// clientVersion reports GoCode's module version, or "dev" for local builds.
func clientVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}
