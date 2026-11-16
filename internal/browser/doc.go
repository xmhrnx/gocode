// Package browser connects GoCode to the user's web browser through the
// GoCode browser extension (browser-extension/ in the repository).
//
// The app listens on a WebSocket at ws://127.0.0.1:<port>/browser. The
// extension connects when the user shares a tab with GoCode from its
// toolbar button, and the browser tool sends it commands over that
// connection. Only extension pages may connect: a web page can't, because
// browsers send the page's own origin with the handshake.
//
// The protocol is JSON, one object per WebSocket message:
//
//	extension → app  {"type":"hello","browser":"Chrome 131"}     once, first
//	app → extension  {"id":1,"method":"snapshot","params":{…}}
//	extension → app  {"id":1,"result":"…"} or {"id":1,"error":"…"}
//	extension → app  {"type":"ping"}   app → extension  {"type":"pong"}
package browser
