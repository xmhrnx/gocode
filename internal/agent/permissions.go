package agent

import (
	"sync"

	"gocode/internal/settings"
	"gocode/internal/tools"
)

// Mode is a chat's overall stance on tool use, picked in the composer.
type Mode string

// Permissions decides whether a tool call needs the user's approval.
type Permissions struct {
	mu            sync.Mutex
	mode          Mode
	policies      map[tools.Access]settings.Policy
	alwaysAllowed map[string]bool
}

const (
	// ModeAsk follows the per-access policies from the settings.
	ModeAsk Mode = "ask"
	// ModeAutoEdit also allows file edits without asking.
	ModeAutoEdit Mode = "auto-edit"
	// ModeFullAccess allows everything without asking.
	ModeFullAccess Mode = "full"
	// ModeReadOnly denies edits and commands: for exploring and planning.
	ModeReadOnly Mode = "read-only"
)

// NewPermissions returns permissions using the policies from the settings.
func NewPermissions(mode Mode, cfg settings.Agent) *Permissions {
	p := &Permissions{mode: mode, alwaysAllowed: map[string]bool{}}
	p.SetPolicies(cfg)
	return p
}

// SetPolicies replaces the per-access policies.
func (p *Permissions) SetPolicies(cfg settings.Agent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.policies = map[tools.Access]settings.Policy{
		tools.AccessRead: cfg.ReadPolicy, tools.AccessWrite: cfg.WritePolicy,
		tools.AccessExecute: cfg.ExecutePolicy, tools.AccessNetwork: cfg.NetworkPolicy,
	}
}

// Mode returns the current mode.
func (p *Permissions) Mode() Mode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mode
}

// SetMode changes the mode; it applies to the next tool call.
func (p *Permissions) SetMode(mode Mode) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mode = mode
}

// AllowAlways stops asking about the tool for the rest of the session.
func (p *Permissions) AllowAlways(toolName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.alwaysAllowed[toolName] = true
}

// Policy returns what to do about a call to tool.
func (p *Permissions) Policy(tool tools.Tool) settings.Policy {
	p.mu.Lock()
	defer p.mu.Unlock()
	access := tool.Access()
	switch {
	case p.mode == ModeReadOnly && (access == tools.AccessWrite || access == tools.AccessExecute):
		return settings.PolicyDeny
	case p.mode == ModeFullAccess:
		return settings.PolicyAllow
	case p.mode == ModeAutoEdit && access == tools.AccessWrite:
		return settings.PolicyAllow
	case p.alwaysAllowed[tool.Spec().Name]:
		return settings.PolicyAllow
	}
	if policy, ok := p.policies[access]; ok {
		return policy
	}
	return settings.PolicyAsk
}
