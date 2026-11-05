package agent

import "gocode/internal/llm"

// loopGuard notices the agent making the same tool call over and over,
// which is how a stuck run looks. It replaces a fixed step limit: long
// tasks run freely, and only a run going in circles is held for the user.
type loopGuard struct {
	lastCall string
	repeats  int
}

// repeatLimit is how many identical calls in a row trigger a check.
const repeatLimit = 3

// observe records call and returns how many times in a row it was made.
func (g *loopGuard) observe(call llm.ToolCall) int {
	signature := call.Name + "\x00" + string(call.Arguments)
	if signature == g.lastCall {
		g.repeats++
	} else {
		g.lastCall, g.repeats = signature, 1
	}
	return g.repeats
}
