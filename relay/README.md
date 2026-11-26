# GoCode relay

A Cloudflare Worker that lets the GoCode phone app reach a desktop on
another network. It only pipes bytes between a phone and a desktop in the
same room: their channel is end-to-end encrypted with the keys they
swapped when pairing, so the relay can't read it, and it needs no
accounts. On the same Wi-Fi the phone connects to the desktop directly
and the relay isn't used.

A desktop connects to its room only while one of its chats has an agent
working (and for two minutes after), so that's when a phone away from
home can reach it.

## Deploy

You need Node.js and a free Cloudflare account. In this folder:

```sh
npx wrangler@latest login
npx wrangler@latest deploy
```

`deploy` prints the address, such as `https://gocode-relay.<you>.workers.dev`.
Put it in GoCode under Settings → Phone → Relay address, then pair the
phone (or pair it again: the address travels in the pairing code).

## Protocol

| Path | Who | What |
| --- | --- | --- |
| `GET /v1/desktop/<room>` | desktop | Control connection. The relay sends `{"type":"incoming","id":"<id>"}` for each phone. |
| `GET /v1/phone/<room>` | phone | Answers 404 when no desktop is in the room; else holds the phone until the desktop takes it. |
| `GET /v1/accept/<room>/<id>` | desktop | Takes phone `<id>`'s stream; the two are piped together. |

All are WebSockets with binary messages. A room is 32 hex characters,
made by the desktop and given to the phone in the pairing code.
