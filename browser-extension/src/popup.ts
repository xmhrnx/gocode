// The toolbar popup: open a tab for GoCode, or share the tab you're on.

import type { PopupRequest, PopupStatus } from "./protocol.js";

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;

const stateText = $("state");
const errorText = $("error");
const openButton = $<HTMLButtonElement>("open");
const shareButton = $<HTMLButtonElement>("share");
const tabList = $<HTMLUListElement>("tabs");
const portInput = $<HTMLInputElement>("port");

let activeTabId: number | undefined;

async function ask(request: PopupRequest): Promise<void> {
  const reply = (await chrome.runtime.sendMessage(request)) as PopupStatus | { failed: string };
  if ("failed" in reply) {
    errorText.textContent = reply.failed;
    return;
  }
  render(reply);
}

function render(status: PopupStatus): void {
  const labels = { off: "Not connected", connecting: "Connecting to GoCode…", connected: "Connected to GoCode" };
  stateText.textContent = labels[status.state];
  stateText.dataset.state = status.state;
  errorText.textContent = status.error;
  if (document.activeElement !== portInput) portInput.value = String(status.port);

  const shared = status.tabs.some((t) => t.id === activeTabId);
  shareButton.textContent = shared ? "Stop sharing this tab" : "Share this tab";
  shareButton.disabled = activeTabId === undefined;
  shareButton.onclick = () => void ask(shared ? { type: "unshare", tabId: activeTabId! } : { type: "share", tabId: activeTabId! });

  tabList.replaceChildren(
    ...status.tabs.map((tab) => {
      const item = document.createElement("li");
      const title = document.createElement("span");
      title.textContent = (tab.current ? "● " : "") + tab.title;
      title.title = tab.url;
      const stop = document.createElement("button");
      stop.textContent = "Stop";
      stop.className = "small";
      stop.onclick = () => void ask({ type: "unshare", tabId: tab.id });
      item.append(title, stop);
      return item;
    }),
  );
  $("shared").hidden = status.tabs.length === 0;
}

openButton.onclick = () => void ask({ type: "open" });

portInput.onchange = () => {
  const port = Number(portInput.value);
  if (Number.isInteger(port) && port > 0 && port < 65536) void ask({ type: "setPort", port });
};

async function start(): Promise<void> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  activeTabId = tab?.id;
  await ask({ type: "status" });
  setInterval(() => void ask({ type: "status" }), 1000);
}

void start();
