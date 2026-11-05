package terminal

// commandHistory remembers typed commands for recall with Up and Down.
type commandHistory struct {
	commands []string
	// cursor is the recalled position; len(commands) means "not recalling".
	cursor int
}

const maxHistory = 500

func (h *commandHistory) add(command string) {
	if n := len(h.commands); n == 0 || h.commands[n-1] != command {
		h.commands = append(h.commands, command)
	}
	if len(h.commands) > maxHistory {
		h.commands = h.commands[len(h.commands)-maxHistory:]
	}
	h.cursor = len(h.commands)
}

// step moves through the history and returns the command there; going
// past the newest returns an empty line.
func (h *commandHistory) step(up bool) (string, bool) {
	switch {
	case up && h.cursor > 0:
		h.cursor--
	case !up && h.cursor < len(h.commands):
		h.cursor++
	default:
		return "", false
	}
	if h.cursor == len(h.commands) {
		return "", true
	}
	return h.commands[h.cursor], true
}
