package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// ToolName is the browser tool's name.
const ToolName = "browser"

// callTimeout bounds one browser command, including page loads.
const callTimeout = 90 * time.Second

// defaultMaxOutputBytes bounds output when Env.MaxOutputBytes is 0.
const defaultMaxOutputBytes = 30000

// Tool lets the agent drive the tabs the user shared through the browser
// extension.
type Tool struct {
	Bridge *Bridge
}

type toolArgs struct {
	Action   string `json:"action"`
	Tab      int    `json:"tab,omitempty"`
	URL      string `json:"url,omitempty"`
	Ref      int    `json:"ref,omitempty"`
	Selector string `json:"selector,omitempty"`
	Text     string `json:"text,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Script   string `json:"script,omitempty"`
}

// actions are the commands the extension understands.
var actions = []string{
	"tabs", "open", "switch", "close", "navigate", "back", "forward", "reload",
	"snapshot", "click", "type", "select", "press", "scroll", "wait", "evaluate",
}

// Spec describes the browser tool to the model.
func (Tool) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: ToolName,
		Description: "Control the user's web browser through the GoCode extension. It works only in tabs the user shared with GoCode " +
			"(by clicking the extension's button) and tabs you open with this tool. If no browser is connected, ask the user to click the GoCode button in their browser's toolbar. " +
			"Start with snapshot: it returns the page's URL, title, visible text and its interactive elements, each with a number ([12] button \"Sign in\"). " +
			"Pass that number as ref to click, type or select; refs change with every snapshot, so take a new one after the page changes. " +
			"Actions: tabs (list shared tabs), open (url, in a new tab), switch (tab), close (tab), navigate (url), back, forward, reload, snapshot, " +
			"click (ref or selector), type (text into ref or selector; key \"Enter\" to submit after typing, value \"append\" to keep existing text), " +
			"select (value of a <select> ref), press (key such as Enter, Escape, Tab, ArrowDown, optionally on ref), " +
			"scroll (text: up, down, top or bottom; or ref to scroll an element into view), wait (text to wait for, or 1-30 seconds as value), " +
			"evaluate (script: the body of an async JavaScript function run in the page; return a JSON-serializable value). " +
			"Commands go to the current tab (the last one opened, switched to or shared) unless tab is given.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "action": {"type": "string", "enum": ["` + strings.Join(actions, `", "`) + `"], "description": "What to do."},
    "tab": {"type": "integer", "description": "Tab id from the tabs action; defaults to the current tab."},
    "url": {"type": "string", "description": "For open and navigate: the address to load."},
    "ref": {"type": "integer", "description": "Element number from the latest snapshot."},
    "selector": {"type": "string", "description": "CSS selector, instead of ref."},
    "text": {"type": "string", "description": "For type: the text to enter. For wait: text to wait for. For scroll: up, down, top or bottom."},
    "key": {"type": "string", "description": "For press: the key. For type: a key to press after typing, e.g. Enter."},
    "value": {"type": "string", "description": "For select: the option's value or label. For wait: seconds. For type: append to keep the existing text."},
    "script": {"type": "string", "description": "For evaluate: JavaScript function body, e.g. return document.title"}
  },
  "required": ["action"],
  "additionalProperties": false
}`),
	}
}

// Access is network: like web_fetch, the tool reaches the web, and the
// Network permission decides whether the user is asked.
func (Tool) Access() tools.Access { return tools.AccessNetwork }

// Summary describes a call in one line, e.g. "Browser: click [12]".
func (Tool) Summary(raw json.RawMessage) string {
	var args toolArgs
	_ = json.Unmarshal(raw, &args)
	parts := []string{"Browser: " + args.Action}
	if args.URL != "" {
		parts = append(parts, args.URL)
	}
	if args.Action == "type" && args.Text != "" {
		parts = append(parts, quote(args.Text))
	}
	if args.Ref > 0 {
		parts = append(parts, fmt.Sprintf("[%d]", args.Ref))
	} else if args.Selector != "" {
		parts = append(parts, args.Selector)
	}
	if args.Key != "" {
		parts = append(parts, args.Key)
	}
	if args.Action == "scroll" || args.Action == "wait" {
		parts = append(parts, args.Text, args.Value)
	}
	return oneLine(strings.Join(parts, " "), 120)
}

// Run sends the command to the extension.
func (t Tool) Run(ctx context.Context, env tools.Env, raw json.RawMessage) tools.Result {
	var args toolArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Errorf("invalid arguments (%v); pass a JSON object matching the tool's parameter schema", err)
	}
	if !contains(actions, args.Action) {
		return tools.Errorf("unknown action %q; use one of %s", args.Action, strings.Join(actions, ", "))
	}
	if t.Bridge == nil {
		return tools.Errorf("%v", ErrNotConnected)
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	result, err := t.Bridge.Call(ctx, args.Action, args)
	if errors.Is(err, ErrNotConnected) {
		return tools.Errorf("No browser is connected. Ask the user to click the GoCode button in their browser's toolbar " +
			"and choose \"Open a new tab\" or \"Share this tab\" (the GoCode browser extension must be installed and GoCode's browser bridge on in Settings → Browser).")
	}
	if err != nil {
		return tools.Errorf("%v", err)
	}
	limit := env.MaxOutputBytes
	if limit <= 0 {
		limit = defaultMaxOutputBytes
	}
	return tools.Result{Output: truncate(resultText(result), limit)}
}

// resultText unwraps a string result; anything else is shown as JSON.
func resultText(result json.RawMessage) string {
	var text string
	if err := json.Unmarshal(result, &text); err == nil {
		if text == "" {
			return "Done."
		}
		return text
	}
	if len(result) == 0 || string(result) == "null" {
		return "Done."
	}
	return string(result)
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && cut < len(s) && s[cut]&0xC0 == 0x80 {
		cut-- // don't split a UTF-8 sequence
	}
	return s[:cut] + fmt.Sprintf("\n… (%d more bytes cut; scroll, or use evaluate to read a specific part)", len(s)-cut)
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if runes := []rune(s); len(runes) > max {
		return string(runes[:max-1]) + "…"
	}
	return s
}

func quote(s string) string {
	return "\"" + oneLine(s, 40) + "\""
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
