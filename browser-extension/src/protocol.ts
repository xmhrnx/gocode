// Messages between the extension and the GoCode app, and between the
// popup and the background service worker.

/** A command from the agent's browser tool. */
export interface Command {
  action: string;
  tab?: number;
  url?: string;
  ref?: number;
  selector?: string;
  text?: string;
  key?: string;
  value?: string;
  script?: string;
}

/** What the app sends: a command to run. */
export interface AppRequest {
  id: number;
  method: string;
  params?: Command;
}

/** Everything the app may send. */
export type AppMessage = AppRequest | { type: "pong" };

/** What the extension sends to the app. */
export type ExtensionMessage =
  | { type: "hello"; browser: string; version: string }
  | { type: "ping" }
  | { id: number; result: string }
  | { id: number; error: string };

export type ConnectionState = "off" | "connecting" | "connected";

export interface SharedTab {
  id: number;
  title: string;
  url: string;
  current: boolean;
}

/** What the popup shows. */
export interface PopupStatus {
  state: ConnectionState;
  error: string;
  port: number;
  tabs: SharedTab[];
}

/** What the popup asks the background to do. */
export type PopupRequest =
  | { type: "status" }
  | { type: "open" }
  | { type: "share"; tabId: number }
  | { type: "unshare"; tabId: number }
  | { type: "setPort"; port: number };
