# GoCode browser extension

Lets the GoCode agent use your browser: open pages, read them, click, type,
pick options and run JavaScript. It works with any model you use in GoCode,
since the app gives every agent the same `browser` tool.

The agent can only touch tabs you share with it:

1. Click the GoCode button in the toolbar.
2. Choose **Open a new tab for GoCode**, or **Share this tab** to hand over
   the tab you're on.

Shared tabs show **ON** on the button. Click **Stop** next to a tab (or close
it) to take it back. When no tab is shared, the extension disconnects.

## Install

You need Node.js 18 or newer.

```bash
npm install
```

```bash
npm run build
```

Then open `chrome://extensions` (or `edge://extensions`), turn on
**Developer mode**, click **Load unpacked** and pick the `dist` folder.
It works in Chrome, Edge, Brave and other Chromium browsers, version 116 or
newer.

## How it connects

GoCode listens on `ws://127.0.0.1:47821/browser` (Settings → Browser, where
it can be turned off or moved to another port; set the same port in the
popup). The connection never leaves your computer, and the app only accepts
browser extensions, so websites can't connect to it.

The agent's commands run in the page with `chrome.scripting`:

| Action | What it does |
| --- | --- |
| `tabs`, `open`, `switch`, `close` | List, open, focus and close shared tabs |
| `navigate`, `back`, `forward`, `reload` | Move around, waiting for the page to load |
| `snapshot` | The page's URL, title, text and numbered interactive elements |
| `click`, `type`, `select`, `press` | Act on an element by its number or a CSS selector |
| `scroll`, `wait` | Scroll the page or wait for text to appear |
| `evaluate` | Run JavaScript in the page and return the result |

Browser pages (`chrome://`, the new tab page, the Web Store) can't be
scripted, so the agent navigates away from them first. Clicks and key presses
are events sent by the page's scripts, which nearly every site accepts; a few
that check for real input (some games, CAPTCHAs) won't react. `evaluate`
doesn't work on sites whose Content Security Policy forbids `eval`.

## Layout

| Path | What lives there |
| --- | --- |
| `src/background.ts` | Service worker: connection lifecycle, popup requests |
| `src/connection.ts` | The WebSocket to GoCode |
| `src/commands.ts` | Runs each action in the shared tabs |
| `src/page.ts` | Functions injected into pages |
| `src/state.ts` | Shared tabs, current tab, port |
| `src/popup.ts`, `static/popup.html` | The toolbar popup |
| `src/protocol.ts` | Message types |
