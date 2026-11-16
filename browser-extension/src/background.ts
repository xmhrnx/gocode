// The service worker: keeps the connection to GoCode open while tabs are
// shared, and answers the popup.

import type { PopupRequest, PopupStatus, SharedTab } from "./protocol.js";
import { connect, connectionStatus, disconnect, errorText, reconnect } from "./connection.js";
import { currentTabId, getPort, markTab, setPort, shareTab, sharedTabIds, unshareTab } from "./state.js";

const KEEPALIVE_ALARM = "gocode-reconnect";

/** Reconnects after the service worker restarts, if tabs are shared. */
async function resume(): Promise<void> {
  if ((await sharedTabIds()).length > 0) await connect();
}

chrome.runtime.onStartup.addListener(() => void resume());
chrome.runtime.onInstalled.addListener(() => void resume());
chrome.alarms.create(KEEPALIVE_ALARM, { periodInMinutes: 0.5 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === KEEPALIVE_ALARM) void resume();
});
void resume();

chrome.tabs.onRemoved.addListener(async (tabId) => {
  if (!(await sharedTabIds()).includes(tabId)) return;
  if ((await unshareTab(tabId)) === 0) disconnect();
});

chrome.tabs.onUpdated.addListener(async (tabId, info) => {
  if (info.status === "loading" && (await sharedTabIds()).includes(tabId)) await markTab(tabId, true);
});

chrome.runtime.onMessage.addListener((request: PopupRequest, _sender, reply) => {
  handlePopup(request).then(reply, (error) => reply({ failed: errorText(error) }));
  return true; // reply comes later
});

async function handlePopup(request: PopupRequest): Promise<PopupStatus> {
  switch (request.type) {
    case "open": {
      const tab = await chrome.tabs.create({ url: chrome.runtime.getURL("welcome.html"), active: true });
      await shareTab(tab.id!);
      await connect();
      break;
    }
    case "share":
      await shareTab(request.tabId);
      await connect();
      break;
    case "unshare":
      if ((await unshareTab(request.tabId)) === 0) disconnect();
      break;
    case "setPort":
      await setPort(request.port);
      await reconnect();
      break;
  }
  return status();
}

async function status(): Promise<PopupStatus> {
  const current = await currentTabId();
  const tabs: SharedTab[] = [];
  for (const id of await sharedTabIds()) {
    const tab = await chrome.tabs.get(id).catch(() => undefined);
    if (tab) tabs.push({ id, title: tab.title || "(no title)", url: tab.url ?? "", current: id === current });
  }
  return { ...connectionStatus(), port: await getPort(), tabs };
}
