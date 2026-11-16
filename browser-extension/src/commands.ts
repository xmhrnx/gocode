// Runs the agent's browser commands in the shared tabs.

import type { Command } from "./protocol.js";
import * as page from "./page.js";
import { currentTabId, setCurrentTab, shareTab, sharedTabIds, unshareTab } from "./state.js";

const MAX_ELEMENTS = 400;
const MAX_TEXT = 20_000;
const LOAD_TIMEOUT_MS = 20_000;

/** Runs one command and returns the text the agent sees. */
export async function runCommand(cmd: Command): Promise<string> {
  switch (cmd.action) {
    case "tabs":
      return listTabs();
    case "open":
      return openTab(requireText(cmd.url, "url"));
    case "switch":
      return switchTab(cmd);
    case "close":
      return closeTab(cmd);
    case "navigate": {
      const tab = await resolveTab(cmd);
      const url = normalizeUrl(requireText(cmd.url, "url"));
      await settle(tab.id!, () => chrome.tabs.update(tab.id!, { url }));
      return describeTab(tab.id!, "Loaded");
    }
    case "back":
    case "forward":
    case "reload": {
      const tab = await resolveTab(cmd);
      const id = tab.id!;
      await settle(id, () => (cmd.action === "reload" ? chrome.tabs.reload(id) : goHistory(id, cmd.action === "back" ? -1 : 1)));
      return describeTab(tab.id!, "Loaded");
    }
    case "snapshot":
      return snapshot(cmd);
    case "click": {
      const tab = await resolveTab(cmd);
      let clicked = "";
      await settle(tab.id!, async () => {
        clicked = (await inPage(tab.id!, page.clickElement, [cmd.ref ?? null, cmd.selector ?? null])).clicked;
      });
      return `Clicked "${clicked}". ${await describeTab(tab.id!, "Now on")}`;
    }
    case "type": {
      const tab = await resolveTab(cmd);
      const result = await inPage(tab.id!, page.typeText, [cmd.ref ?? null, cmd.selector ?? null, cmd.text ?? "", cmd.value === "append"]);
      let reply = `Typed. The field now holds: ${JSON.stringify(result.value)}`;
      if (cmd.key) {
        await settle(tab.id!, () => inPage(tab.id!, page.pressKey, [cmd.ref ?? null, cmd.selector ?? null, cmd.key!]));
        reply += `\nPressed ${cmd.key}. ${await describeTab(tab.id!, "Now on")}`;
      }
      return reply;
    }
    case "select": {
      const tab = await resolveTab(cmd);
      const result = await inPage(tab.id!, page.selectOption, [cmd.ref ?? null, cmd.selector ?? null, requireText(cmd.value, "value")]);
      return `Selected "${result.selected}".`;
    }
    case "press": {
      const tab = await resolveTab(cmd);
      const key = requireText(cmd.key, "key");
      await settle(tab.id!, () => inPage(tab.id!, page.pressKey, [cmd.ref ?? null, cmd.selector ?? null, key]));
      return `Pressed ${key}. ${await describeTab(tab.id!, "Now on")}`;
    }
    case "scroll": {
      const tab = await resolveTab(cmd);
      const result = await inPage(tab.id!, page.scrollPage, [cmd.ref ?? null, cmd.selector ?? null, cmd.text || "down"]);
      return `Scrolled to ${result.position}. Take a snapshot to see what's there.`;
    }
    case "wait":
      return wait(cmd);
    case "evaluate": {
      const tab = await resolveTab(cmd);
      const result = await inPage(tab.id!, page.evaluateScript, [requireText(cmd.script, "script")], "MAIN");
      return result.value;
    }
    default:
      throw new Error(`Unknown action "${cmd.action}".`);
  }
}

async function listTabs(): Promise<string> {
  const ids = await sharedTabIds();
  if (ids.length === 0) return "No tabs are shared. Use open to open one.";
  const current = await currentTabId();
  const lines: string[] = [];
  for (const id of ids) {
    try {
      const tab = await chrome.tabs.get(id);
      lines.push(`${id === current ? "* " : "  "}tab ${id}: ${tab.title || "(no title)"} — ${tab.url ?? ""}`);
    } catch {
      await unshareTab(id);
    }
  }
  return `Shared tabs (* is current):\n${lines.join("\n")}`;
}

async function openTab(url: string): Promise<string> {
  const target = normalizeUrl(url);
  const tab = await chrome.tabs.create({ url: "about:blank", active: true });
  await shareTab(tab.id!);
  await settle(tab.id!, () => chrome.tabs.update(tab.id!, { url: target }));
  return describeTab(tab.id!, `Opened tab ${tab.id}:`);
}

async function switchTab(cmd: Command): Promise<string> {
  if (cmd.tab === undefined) throw new Error("Give the tab to switch to (see the tabs action).");
  const tab = await resolveTab(cmd);
  await chrome.tabs.update(tab.id!, { active: true });
  await chrome.windows.update(tab.windowId, { focused: true });
  return describeTab(tab.id!, `Switched to tab ${tab.id}:`);
}

async function closeTab(cmd: Command): Promise<string> {
  const tab = await resolveTab(cmd);
  await unshareTab(tab.id!);
  await chrome.tabs.remove(tab.id!);
  const next = await currentTabId();
  return `Closed tab ${tab.id}.` + (next !== undefined ? ` The current tab is now ${next}.` : " No shared tabs are left.");
}

async function snapshot(cmd: Command): Promise<string> {
  const tab = await resolveTab(cmd);
  const s = await inPage(tab.id!, page.snapshotPage, [MAX_ELEMENTS, MAX_TEXT]);
  const parts = [
    `Tab ${tab.id}: ${s.title || "(no title)"}`,
    `URL: ${s.url}`,
    `Scrolled ${s.scrollY} of ${s.scrollHeight}px (window ${s.viewportHeight}px)`,
    "",
    "Interactive elements:",
    s.elements.length ? s.elements.join("\n") : "(none)",
  ];
  if (s.moreElements) parts.push(`… ${s.moreElements} more; scroll and snapshot again, or use a selector.`);
  parts.push("", "Page text:", s.text || "(none)");
  return parts.join("\n");
}

async function wait(cmd: Command): Promise<string> {
  const tab = await resolveTab(cmd);
  const seconds = Math.min(Math.max(Number(cmd.value) || (cmd.text ? 10 : 2), 0.1), 30);
  if (!cmd.text) {
    await sleep(seconds * 1000);
    return `Waited ${seconds}s. ${await describeTab(tab.id!, "Now on")}`;
  }
  const result = await inPage(tab.id!, page.waitForText, [cmd.text, seconds * 1000]);
  return result.found ? `"${cmd.text}" is on the page.` : `"${cmd.text}" didn't appear within ${seconds}s.`;
}

/** The tab a command is for: the given one or the current one, if shared. */
async function resolveTab(cmd: Command): Promise<chrome.tabs.Tab> {
  const shared = await sharedTabIds();
  if (shared.length === 0) {
    throw new Error("No tab is shared with GoCode. Use open to open one, or ask the user to share a tab from the GoCode extension button.");
  }
  const id = cmd.tab ?? (await currentTabId()) ?? shared[shared.length - 1];
  if (!shared.includes(id)) {
    throw new Error(`Tab ${id} isn't shared with GoCode. Shared tabs: ${shared.join(", ")}.`);
  }
  try {
    const tab = await chrome.tabs.get(id);
    if (cmd.tab !== undefined) await setCurrentTab(id);
    return tab;
  } catch {
    await unshareTab(id);
    throw new Error(`Tab ${id} was closed.`);
  }
}

/** Runs a page function in a tab and unwraps its result. */
async function inPage<Args extends unknown[], T>(
  tabId: number,
  func: (...args: Args) => page.PageResult<T> | Promise<page.PageResult<T>>,
  args: Args,
  world: chrome.scripting.ExecutionWorld = "ISOLATED",
): Promise<T> {
  let result: page.PageResult<T> | undefined;
  try {
    const results = await chrome.scripting.executeScript({ target: { tabId }, func, args, world });
    result = results[0]?.result as page.PageResult<T> | undefined;
  } catch (error) {
    const tab = await chrome.tabs.get(tabId).catch(() => undefined);
    throw new Error(
      `Can't reach the page ${tab?.url ?? ""} (${error instanceof Error ? error.message : String(error)}). ` +
        "Browser pages (chrome://, new tab, extension and store pages) can't be controlled; navigate to a website first.",
    );
  }
  if (!result) throw new Error("The page didn't answer; it may have navigated away. Take a new snapshot.");
  if (!result.ok) throw new Error(result.error);
  return result;
}

/**
 * Runs action, then waits for any page load it started: a navigation
 * starts within a moment, so if the tab is still idle after that, there
 * was none.
 */
async function settle(tabId: number, action: () => Promise<unknown>, quietMs = 500): Promise<void> {
  let loading = false;
  let loaded!: () => void;
  const done = new Promise<void>((resolve) => (loaded = resolve));
  const listener = (id: number, info: chrome.tabs.TabChangeInfo) => {
    if (id !== tabId) return;
    if (info.status === "loading") loading = true;
    if (info.status === "complete" && loading) loaded();
  };
  const removed = (id: number) => id === tabId && loaded();
  chrome.tabs.onUpdated.addListener(listener);
  chrome.tabs.onRemoved.addListener(removed);
  try {
    await action();
    await sleep(quietMs);
    const tab = await chrome.tabs.get(tabId).catch(() => undefined);
    if (tab?.status === "loading") loading = true;
    if (loading) await Promise.race([done, sleep(LOAD_TIMEOUT_MS)]);
  } finally {
    chrome.tabs.onUpdated.removeListener(listener);
    chrome.tabs.onRemoved.removeListener(removed);
  }
}

/**
 * Moves through the tab's history with the page's own history.go: Chrome's
 * back button (and chrome.tabs.goBack) skips pages that were left without
 * a user gesture, which is every page the agent clicked away from.
 */
async function goHistory(tabId: number, delta: number): Promise<void> {
  try {
    await chrome.scripting.executeScript({ target: { tabId }, func: (d: number) => history.go(d), args: [delta] });
  } catch {
    await (delta < 0 ? chrome.tabs.goBack(tabId) : chrome.tabs.goForward(tabId)); // a page scripts can't reach
  }
}

async function describeTab(tabId: number, prefix: string): Promise<string> {
  const tab = await chrome.tabs.get(tabId).catch(() => undefined);
  if (!tab) return "The tab was closed.";
  return `${prefix} ${tab.title || "(no title)"} — ${tab.url ?? ""}`;
}

function normalizeUrl(url: string): string {
  const trimmed = url.trim();
  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed) || /^(about|data|javascript|file):/i.test(trimmed)) return trimmed;
  const local = /^(localhost|127\.|\[::1\]|0\.0\.0\.0)/i.test(trimmed);
  return `${local ? "http" : "https"}://${trimmed}`;
}

function requireText(value: string | undefined, name: string): string {
  if (!value) throw new Error(`This action needs ${name}.`);
  return value;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
