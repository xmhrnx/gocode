// GoCode relay: lets the GoCode phone app reach a desktop that is on
// another network. It only pipes bytes between a phone and a desktop in
// the same room; their channel is end-to-end encrypted, so the relay can't
// read or change it. One Durable Object holds each room.
//
//   GET /v1/desktop/<room>      the desktop's control connection, held
//                               while its agent works
//   GET /v1/phone/<room>        a phone; 404 when no desktop is there
//   GET /v1/accept/<room>/<id>  the desktop taking phone <id>'s stream

const ROOM = /^[0-9a-f]{32}$/;
const ID = /^[0-9a-f]{16}$/;
// How long a phone may wait for the desktop to take its stream.
const ACCEPT_WAIT_MS = 15000;
// The most a phone may send before the desktop takes its stream.
const MAX_BUFFERED = 1 << 20;

export default {
  async fetch(request, env) {
    const parts = new URL(request.url).pathname.split("/").filter(Boolean);
    if (parts[0] !== "v1" || !["desktop", "phone", "accept"].includes(parts[1]) || !ROOM.test(parts[2] || "")) {
      return new Response("GoCode relay", { status: 404 });
    }
    if (request.headers.get("Upgrade") !== "websocket") {
      return new Response("expected a WebSocket", { status: 426 });
    }
    const room = env.ROOMS.get(env.ROOMS.idFromName(parts[2]));
    return room.fetch(request);
  },
};

export class Room {
  constructor(state) {
    this.state = state;
    this.desktop = null;
    // Phones waiting for the desktop to take their stream, by id.
    this.waiting = new Map();
  }

  async fetch(request) {
    const parts = new URL(request.url).pathname.split("/").filter(Boolean);
    switch (parts[1]) {
      case "desktop":
        return this.connectDesktop();
      case "phone":
        return this.connectPhone();
      case "accept":
        return this.accept(parts[3] || "");
    }
    return new Response("not found", { status: 404 });
  }

  connectDesktop() {
    const [client, server] = Object.values(new WebSocketPair());
    server.accept();
    if (this.desktop) {
      try { this.desktop.close(1000, "replaced"); } catch {}
    }
    this.desktop = server;
    const gone = () => {
      if (this.desktop === server) this.desktop = null;
    };
    server.addEventListener("close", gone);
    server.addEventListener("error", gone);
    return new Response(null, { status: 101, webSocket: client });
  }

  connectPhone() {
    if (!this.desktop) {
      return new Response("the desktop is offline", { status: 404 });
    }
    const [client, server] = Object.values(new WebSocketPair());
    server.accept();
    const id = randomID();
    const entry = { socket: server, buffered: [], size: 0, peer: null };
    entry.timer = setTimeout(() => this.drop(id, 4408, "the desktop didn't answer"), ACCEPT_WAIT_MS);
    this.waiting.set(id, entry);
    server.addEventListener("message", (event) => {
      if (entry.peer) {
        entry.peer.send(event.data);
        return;
      }
      entry.size += byteLength(event.data);
      if (entry.size > MAX_BUFFERED) {
        this.drop(id, 1009, "too much data before the desktop answered");
        return;
      }
      entry.buffered.push(event.data);
    });
    server.addEventListener("close", () => this.closeBoth(entry));
    server.addEventListener("error", () => this.closeBoth(entry));
    try {
      this.desktop.send(JSON.stringify({ type: "incoming", id }));
    } catch {
      this.drop(id, 1011, "the desktop is offline");
    }
    return new Response(null, { status: 101, webSocket: client });
  }

  accept(id) {
    const entry = ID.test(id) ? this.waiting.get(id) : undefined;
    if (!entry) {
      return new Response("no such phone", { status: 404 });
    }
    this.waiting.delete(id);
    clearTimeout(entry.timer);
    const [client, server] = Object.values(new WebSocketPair());
    server.accept();
    entry.peer = server;
    for (const data of entry.buffered) server.send(data);
    entry.buffered = [];
    server.addEventListener("message", (event) => {
      try { entry.socket.send(event.data); } catch {}
    });
    server.addEventListener("close", () => this.closeBoth(entry));
    server.addEventListener("error", () => this.closeBoth(entry));
    return new Response(null, { status: 101, webSocket: client });
  }

  drop(id, code, reason) {
    const entry = this.waiting.get(id);
    if (!entry) return;
    this.waiting.delete(id);
    try { entry.socket.close(code, reason); } catch {}
  }

  closeBoth(entry) {
    clearTimeout(entry.timer);
    for (const [id, e] of this.waiting) {
      if (e === entry) this.waiting.delete(id);
    }
    for (const socket of [entry.socket, entry.peer]) {
      if (!socket) continue;
      try { socket.close(1000, "closed"); } catch {}
    }
  }
}

function randomID() {
  const bytes = crypto.getRandomValues(new Uint8Array(8));
  return [...bytes].map((b) => b.toString(16).padStart(2, "0")).join("");
}

function byteLength(data) {
  return typeof data === "string" ? data.length : data.byteLength;
}
