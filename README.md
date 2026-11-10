# GoCode

A native coding agent for the desktop. It's written in Go with
[Giode](https://github.com/xmhrnx/giode) (on Gio), so there's no browser engine,
webview or JavaScript runtime inside. The UI is drawn straight to the GPU.

## Features

- **Tabs** in a custom title bar: agent chats and settings.
  Drag tabs to reorder them, right-click one to close or move it, or move the
  current tab with Ctrl/⌘+Shift+Page Up / Page Down.
- **Agent loop**: the model streams a reply, calls tools, sees their results and
  continues until the task is done. There is no step limit by default. A run
  that repeats the same tool call 3 times in a row pauses for your approval.
  Failed requests are retried with backoff. Esc stops a run, and the agent's
  plan shows above the composer.
- **Messages while it works**: send messages while the agent is running.
  They're queued and reach the model at its next step, with no need to stop
  it first; Send now stops the current step and sends them at once.
- **Project folders**: open a folder for a chat, or just start typing: a chat
  without one gets its own folder in `~/GoCode` (Settings → General →
  Projects folder), named after its first message.
- **Tabs reopen** where you left off: chats and settings
  (Settings → General).
- **Shell**: the agent's commands run in `$SHELL`. On Windows they run in Git Bash when
  it's installed, then PowerShell 7, then Windows PowerShell. The model is
  told which shell it has.
- **Sub-agents**: the `task` tool hands a self-contained job (a broad search,
  a review, an investigation) to a sub-agent with a fresh context and the same
  tools and permissions. Only its report reaches the main conversation, which
  keeps that conversation's context small. Its steps show live in the task's
  card. It can be switched off in Settings → Permissions.
- **Context compaction**: when the conversation nears the model's context
  window, or the provider rejects it as too long, GoCode summarizes it and
  carries on. The transcript stays visible, the header shows how full the
  context is, and a button compacts on demand.
- **MCP servers**: tools from Model Context Protocol servers, run as a program
  (stdio) or reached at a URL (Streamable HTTP). Add them in Settings → MCP
  servers.
- **Skills**: Agent Skills folders, each with a `SKILL.md`. Install them from a
  folder, a `.zip`, a GitHub repository or a zip URL. Project skills in
  `.gocode/skills` or `.claude/skills` are found too. The agent sees each
  skill's name and description and loads the full skill with the `skill` tool.
- **Tools**: `read_file`, `write_file`, `edit_file` (exact replacement with a
  diff), `list_dir`, `glob`, `grep`, `bash` (with timeout and process-group
  kill), `web_fetch`, `web_search` and `todo_write`.
- **Web search**: the `web_search` tool uses Brave Search, Tavily, Exa or your
  own SearXNG (Settings → Web search). Anthropic, Gemini 3 and OpenRouter
  models can use their provider's own search instead, and the MCP server
  form has presets for the Brave, Tavily and Exa MCP servers.
- **Permissions**: an inline approval card for each tool call, showing the diff
  for edits. It offers Allow, Always allow (for the session) and Deny with
  feedback for the agent. Each chat has a mode: *Ask*, *Auto-accept edits*,
  *Full access*, *Read-only (plan)* or *Chat only*.
- **Providers**: 34 built-in, plus any custom endpoint.
  - Native APIs: Anthropic Messages, Google Gemini.
  - OpenAI-compatible: OpenAI, xAI, DeepSeek, Mistral, Moonshot, Z.ai, Qwen,
    MiniMax, Cohere, OpenRouter, Groq, Together, Fireworks, Cerebras, DeepInfra,
    Hugging Face, Nebius, SambaNova, Novita, NVIDIA, Perplexity, Azure OpenAI,
    Bedrock and Ollama Cloud.
  - Local servers: LM Studio, Ollama, llama.cpp, vLLM, Jan, LocalAI, KoboldCpp
    and text-generation-webui.
  - Custom providers can speak any of the three API styles.
- **Model picker** on every chat (Ctrl/⌘+K): recent models and the connected
  providers. Each provider opens its own searchable model list. Models you
  don't use can be hidden in Settings → Models.
- **Attachments**: files and images through the paperclip button, by dropping
  them on the window, or pasted from the clipboard (Ctrl/⌘+V or the image
  button). PDFs go to Anthropic, Gemini, OpenAI, OpenRouter and Azure models
  as documents, and as their text to the rest. Word, Excel, PowerPoint,
  OpenDocument and EPUB files are sent as text, as are text and code files
  (UTF-8 or UTF-16). Images (PNG, JPEG, GIF, WebP, BMP, TIFF) are scaled
  down before sending. Audio (WAV, MP3, AAC, FLAC, Ogg, AIFF, M4A) and video
  (MP4, MOV, WebM, AVI, MPEG, FLV, WMV, 3GP) go to Gemini whole, WAV and MP3
  to OpenAI and OpenRouter audio models; other models get a description of
  the file and where it is, so the agent can work on it with its tools.
- **Selecting text**: drag across the conversation to select
  text, even across messages or blocks, and copy it with Ctrl/⌘+C.
- **Messages**: copy, retry, and edit-and-resend. Replies render as Markdown
  with code blocks you can copy. Reasoning is shown collapsed.
- **Settings** are grouped into categories and searchable. They include the send
  key (Enter, or Ctrl+Enter so plain Enter never sends), 9 themes plus System, text sizes,
  editable system prompts, per-tool switches and network retries.
- **History**: chats are saved as JSON and can be reopened from the history
  dialog (Ctrl/⌘+H).
- **API keys** are stored in the system keychain (macOS Keychain, Windows DPAPI,
  Linux Secret Service). When no keychain is available they go in a file
  readable only by you.

## Build

You need Go 1.26 or newer. On Linux, Gio also needs the Wayland, X11, EGL and
Vulkan headers, for example on Debian or Ubuntu:

```bash
sudo apt install libwayland-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libx11-xcb-dev libxcursor-dev libvulkan-dev libxfixes-dev
```

To build:

```bash
go build -o gocode .
```

The [Giode CLI](https://github.com/xmhrnx/giode-cli) builds packages
(`giode build windows-installer`, `dmg`, `deb`, `appimage`, …) from
`giode.toml`. It also builds against Giode's patched Gio, which you need for
dropping files onto the window on Linux.

## Releases and website

Publishing a GitHub release (tag `vX.Y.Z`) runs `.github/workflows/release.yml`,
written by `giode ci` from the `[ci]` section of `giode.toml`; rerun `giode ci -f`
after changing it. It tests, then builds Windows installers and MSIX packages
(x64 and ARM64), Linux `.deb` and `.rpm` packages (x64 and ARM64) and a universal
macOS `.dmg`, and attaches them to the release.

The website in `website/` is published to GitHub Pages by the `website` workflow,
run by hand from the Actions tab. Its download button picks the file for the
visitor's OS and CPU from the latest release.

## Layout

| Path | What lives there |
| --- | --- |
| `main.go` | Wiring only |
| `assets/` | Embedded fonts (Inter, JetBrains Mono), Lucide icons, app icon |
| `internal/llm` | Provider-neutral messages, requests, streaming events |
| `internal/llm/{openai,anthropic,gemini,sse}` | API adapters |
| `internal/providers` | Provider catalog and client factory |
| `internal/agent` | Agent loop, permissions, retries, compaction, repeat guard |
| `internal/mcp` | MCP client (stdio and Streamable HTTP) and its tools |
| `internal/skills` | Skill discovery, installation and the `skill` tool |
| `internal/tools` | The agent's tools |
| `internal/websearch` | Search services behind the `web_search` tool |
| `internal/prompts` | System prompts (defaults in `defaults/*.md`) |
| `internal/{settings,session,secrets,catalog,attach,shell}` | Preferences, chat storage, keys, model lists, attachments, the agent's shell |
| `ui/` | Window root, shortcuts; one component per file in `chat/`, `settingsview/`, `modelpicker/`, `history/`, `workspace/`, `components/`, `theme/` |

## Project instructions

When "Follow project instructions" is on (the default), the agent reads
`AGENTS.md`, `CLAUDE.md` or `.gocode/rules.md` from the project folder. Your own
rules for every project go in Settings → Prompts.
