// The tabs shared with GoCode, the current tab and the port. Shared tabs
// live in session storage, which outlives the service worker but not the
// browser, like tab ids themselves.

export const DEFAULT_PORT = 47821;

const SHARED = "sharedTabs";
const CURRENT = "currentTab";
const PORT = "port";

let lock: Promise<unknown> = Promise.resolve();

/** Runs fn after every earlier update, so read-modify-writes don't race. */
function serialized<T>(fn: () => Promise<T>): Promise<T> {
  const next = lock.then(fn, fn);
  lock = next.catch(() => undefined);
  return next;
}

export async function sharedTabIds(): Promise<number[]> {
  const stored = await chrome.storage.session.get(SHARED);
  return (stored[SHARED] as number[] | undefined) ?? [];
}

export async function currentTabId(): Promise<number | undefined> {
  const stored = await chrome.storage.session.get(CURRENT);
  return stored[CURRENT] as number | undefined;
}

export function setCurrentTab(tabId: number): Promise<void> {
  return chrome.storage.session.set({ [CURRENT]: tabId });
}

/** Shares a tab and makes it the current one. */
export function shareTab(tabId: number): Promise<void> {
  return serialized(async () => {
    const ids = await sharedTabIds();
    if (!ids.includes(tabId)) ids.push(tabId);
    await chrome.storage.session.set({ [SHARED]: ids, [CURRENT]: tabId });
    await markTab(tabId, true);
  });
}

/** Stops sharing a tab; returns how many tabs are still shared. */
export function unshareTab(tabId: number): Promise<number> {
  return serialized(async () => {
    const ids = (await sharedTabIds()).filter((id) => id !== tabId);
    const current = await currentTabId();
    await chrome.storage.session.set({ [SHARED]: ids });
    if (current === tabId) {
      if (ids.length > 0) await setCurrentTab(ids[ids.length - 1]);
      else await chrome.storage.session.remove(CURRENT);
    }
    await markTab(tabId, false);
    return ids.length;
  });
}

/** Shows "ON" on the toolbar button while a shared tab is in front. */
export async function markTab(tabId: number, shared: boolean): Promise<void> {
  try {
    await chrome.action.setBadgeText({ tabId, text: shared ? "ON" : "" });
    if (shared) await chrome.action.setBadgeBackgroundColor({ tabId, color: "#2563eb" });
  } catch {
    // The tab is gone.
  }
}

export async function getPort(): Promise<number> {
  const stored = await chrome.storage.local.get(PORT);
  const port = stored[PORT] as number | undefined;
  return port && port > 0 && port < 65536 ? port : DEFAULT_PORT;
}

export function setPort(port: number): Promise<void> {
  return chrome.storage.local.set({ [PORT]: port });
}
