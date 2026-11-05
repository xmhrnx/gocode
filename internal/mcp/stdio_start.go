package mcp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"gocode/internal/settings"
)

// startStdio starts the server's command and its reader goroutine.
func startStdio(server settings.MCPServer, events serverEvents) (transport, error) {
	if server.Command == "" {
		return nil, errors.New("stdio MCP server has no command")
	}
	cmd := exec.Command(server.Command, server.Args...)
	cmd.Env = append(os.Environ(), envList(server.Env)...)
	cmd.WaitDelay = pipeDrainDelay
	startInOwnGroup(cmd)
	t := &stdioTransport{cmd: cmd, stderr: newTailBuffer(stderrTailBytes), events: events,
		pending: make(map[string]chan message), done: make(chan struct{})}
	cmd.Stderr = t.stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", server.Command, err)
	}
	t.stdin = stdin
	go t.readLoop(stdout)
	return t, nil
}

// envList turns env into sorted KEY=value entries for exec.Cmd.
func envList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for key, value := range env {
		list = append(list, key+"="+value)
	}
	sort.Strings(list)
	return list
}
