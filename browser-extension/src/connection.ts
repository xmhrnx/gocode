// The WebSocket to the GoCode app. It is open while at least one tab is
// shared: the app sends commands, and the extension runs them and answers.

import type { AppMessage, ConnectionState, ExtensionMessage } from "./protocol.js";
import { runCommand } from "./commands.js";
import { getPort, sharedTabIds } from "./state.js";

const PING_MS = 20_000; // keeps the service worker alive (Chrome 116+)
const RETRY_MS = 3_000;

let socket: WebSocket | undefined;
let state: ConnectionState = "off";
let lastError = "";
let pingTimer: ReturnType<typeof setInterval> | undefined;
let retryTimer: ReturnType<typeof setTimeout> | undefined;

export function connectionStatus(): { state: ConnectionState; error: string } {
  return { state, error: lastError };
}

/** Connects to the app unless already connected or connecting. */
export async function connect(): Promise<void> {
  clearTimeout(retryTimer);
  if (socket) return;
  const port = await getPort();
  if (socket) return; // connected while reading the port
  const ws = new WebSocket(`ws://127.0.0.1:${port}/browser`);
  socket = ws;
  state = "connecting";
  ws.onopen = () => {
    state = "connected";
    lastError = "";
    send(ws, { type: "hello", browser: browserName(), version: chrome.runtime.getManifest().version });
    pingTimer = setInterval(() => send(ws, { type: "ping" }), PING_MS);
  };
  ws.onmessage = (event) => void handle(ws, String(event.data));
  ws.onclose = () => {
    if (socket !== ws) return;
    if (state !== "connected") {
      lastError = `GoCode isn't reachable on port ${port}. Is the app running, with Settings → Browser on?`;
    }
    socket = undefined;
    state = "off";
    clearInterval(pingTimer);
    void retryIfShared();
  };
}

/** Closes the connection, e.g. when no tab is shared any more. */
export function disconnect(): void {
  clearTimeout(retryTimer);
  clearInterval(pingTimer);
  const ws = socket;
  socket = undefined;
  state = "off";
  lastError = "";
  ws?.close();
}

/** Reconnects to a new port. */
export async function reconnect(): Promise<void> {
  disconnect();
  if ((await sharedTabIds()).length > 0) await connect();
}

async function retryIfShared(): Promise<void> {
  if ((await sharedTabIds()).length > 0) {
    retryTimer = setTimeout(() => void connect(), RETRY_MS);
  }
}

async function handle(ws: WebSocket, data: string): Promise<void> {
  let message: AppMessage;
  try {
    message = JSON.parse(data) as AppMessage;
  } catch {
    return;
  }
  if (!("id" in message)) return; // pong
  const command = { ...message.params, action: message.params?.action ?? message.method };
  try {
    send(ws, { id: message.id, result: await runCommand(command) });
  } catch (error) {
    send(ws, { id: message.id, error: errorText(error) });
  }
}

function send(ws: WebSocket, message: ExtensionMessage): void {
  if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(message));
}

export function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

interface UserAgentData {
  brands?: { brand: string; version: string }[];
}

/** Names the browser, e.g. "Google Chrome 131". */
function browserName(): string {
  const data = (navigator as Navigator & { userAgentData?: UserAgentData }).userAgentData;
  const brands = data?.brands?.filter((b) => !/not.?a.?brand/i.test(b.brand)) ?? [];
  const brand = brands.find((b) => b.brand !== "Chromium") ?? brands[0];
  return brand ? `${brand.brand} ${brand.version}` : navigator.userAgent;
}
