// Functions run inside the page with chrome.scripting.executeScript. Each
// is serialized on its own, so it must not use anything from outside its
// body: helpers are declared inside the function that needs them. They
// return { error } instead of throwing, since thrown errors don't reach
// the extension.

export type PageResult<T> = ({ ok: true } & T) | { ok?: false; error: string };

export interface Snapshot {
  url: string;
  title: string;
  scrollY: number;
  scrollHeight: number;
  viewportHeight: number;
  elements: string[];
  moreElements: number;
  text: string;
}

/** Numbers the page's interactive elements and reads its text. */
export function snapshotPage(maxElements: number, maxText: number): PageResult<Snapshot> {
  const selector = [
    "a[href]", "button", "input:not([type=hidden])", "select", "textarea", "summary",
    "[role=button]", "[role=link]", "[role=checkbox]", "[role=radio]", "[role=switch]", "[role=tab]",
    "[role=menuitem]", "[role=option]", "[role=combobox]", "[role=textbox]", "[role=searchbox]",
    "[contenteditable='']", "[contenteditable=true]", "[onclick]",
  ].join(",");
  const clean = (s: string | null | undefined, max = 80): string => {
    const t = (s ?? "").replace(/\s+/g, " ").trim();
    return t.length > max ? t.slice(0, max - 1) + "…" : t;
  };
  const visible = (el: HTMLElement): boolean => {
    const rect = el.getBoundingClientRect();
    if (rect.width === 0 && rect.height === 0) return false;
    const style = getComputedStyle(el);
    return style.visibility !== "hidden" && style.display !== "none" && !el.closest("[aria-hidden=true]");
  };
  const label = (el: HTMLElement): string => {
    const field = el as HTMLInputElement;
    const isField = el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement || el instanceof HTMLSelectElement;
    const labelledBy = el.getAttribute("aria-labelledby");
    return clean(
      el.getAttribute("aria-label") ||
        (labelledBy ? document.getElementById(labelledBy)?.innerText : "") ||
        (isField ? field.labels?.[0]?.innerText : "") ||
        el.getAttribute("placeholder") ||
        (isField ? "" : el.innerText) ||
        el.getAttribute("title") ||
        el.querySelector("img[alt]")?.getAttribute("alt") ||
        el.getAttribute("alt") ||
        el.getAttribute("name") ||
        "",
    );
  };
  const describe = (el: HTMLElement): string => {
    const tag = el.tagName.toLowerCase();
    const parts: string[] = [el.getAttribute("role") ?? (tag === "a" ? "link" : tag)];
    if (el instanceof HTMLInputElement) parts.push(`type=${el.type}`);
    const name = label(el);
    if (name) parts.push(JSON.stringify(name));
    if (el instanceof HTMLInputElement && (el.type === "checkbox" || el.type === "radio")) {
      parts.push(el.checked ? "checked" : "unchecked");
    } else if (el instanceof HTMLInputElement && el.type === "password") {
      if (el.value) parts.push("value=(hidden)");
    } else if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
      if (el.value) parts.push(`value=${JSON.stringify(clean(el.value, 120))}`);
    } else if (el instanceof HTMLSelectElement) {
      parts.push(`selected=${JSON.stringify(clean(el.selectedOptions[0]?.text))}`);
      const options = Array.from(el.options).slice(0, 15).map((o) => clean(o.text, 40));
      parts.push(`options: ${options.join(" | ")}${el.options.length > 15 ? " | …" : ""}`);
    } else if (el.getAttribute("aria-checked")) {
      parts.push(el.getAttribute("aria-checked") === "true" ? "checked" : "unchecked");
    } else if (el.getAttribute("aria-expanded")) {
      parts.push(el.getAttribute("aria-expanded") === "true" ? "expanded" : "collapsed");
    }
    if (el instanceof HTMLAnchorElement) parts.push(`-> ${clean(el.getAttribute("href"), 100)}`);
    if ((el as HTMLButtonElement).disabled || el.getAttribute("aria-disabled") === "true") parts.push("disabled");
    const rect = el.getBoundingClientRect();
    if (rect.bottom < 0 || rect.top > innerHeight) parts.push("(off screen)");
    return parts.join(" ");
  };

  for (const old of document.querySelectorAll("[data-gocode-ref]")) old.removeAttribute("data-gocode-ref");
  const elements: string[] = [];
  let more = 0;
  for (const el of document.querySelectorAll<HTMLElement>(selector)) {
    if (!visible(el)) continue;
    if (elements.length >= maxElements) {
      more++;
      continue;
    }
    const ref = elements.length + 1;
    el.setAttribute("data-gocode-ref", String(ref));
    elements.push(`[${ref}] ${describe(el)}`);
  }
  const text = (document.body?.innerText ?? "").replace(/[ \t]+/g, " ").replace(/\n\s*\n+/g, "\n").trim();
  const root = document.scrollingElement ?? document.documentElement;
  return {
    ok: true,
    url: location.href,
    title: document.title,
    scrollY: Math.round(scrollY),
    scrollHeight: root.scrollHeight,
    viewportHeight: innerHeight,
    elements,
    moreElements: more,
    text: text.length > maxText ? text.slice(0, maxText) + "\n… (more text below; scroll or use evaluate)" : text,
  };
}

/** Clicks an element the way a mouse would. */
export function clickElement(ref: number | null, selector: string | null): PageResult<{ clicked: string }> {
  const el = (ref ? document.querySelector(`[data-gocode-ref="${ref}"]`) : selector ? document.querySelector(selector) : null) as HTMLElement | null;
  if (!el) return { error: ref ? `There's no element [${ref}] now; take a new snapshot.` : `No element matches ${selector}.` };
  el.scrollIntoView({ block: "center", inline: "center" });
  const rect = el.getBoundingClientRect();
  const init: MouseEventInit = {
    bubbles: true, cancelable: true, composed: true, view: window, button: 0,
    clientX: rect.left + rect.width / 2, clientY: rect.top + rect.height / 2,
  };
  el.dispatchEvent(new PointerEvent("pointerdown", { ...init, pointerType: "mouse", isPrimary: true }));
  el.dispatchEvent(new MouseEvent("mousedown", init));
  el.focus({ preventScroll: true });
  el.dispatchEvent(new PointerEvent("pointerup", { ...init, pointerType: "mouse", isPrimary: true }));
  el.dispatchEvent(new MouseEvent("mouseup", init));
  el.click();
  const name = (el.getAttribute("aria-label") || el.innerText || el.getAttribute("value") || el.tagName.toLowerCase()).replace(/\s+/g, " ").trim();
  return { ok: true, clicked: name.slice(0, 80) };
}

/** Types text into a field, replacing what's there unless append. */
export function typeText(ref: number | null, selector: string | null, text: string, append: boolean): PageResult<{ value: string }> {
  const el = (ref ? document.querySelector(`[data-gocode-ref="${ref}"]`) : selector ? document.querySelector(selector) : document.activeElement) as HTMLElement | null;
  if (!el) return { error: ref ? `There's no element [${ref}] now; take a new snapshot.` : `No element matches ${selector}.` };
  el.scrollIntoView({ block: "center" });
  el.focus();
  if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
    if (append) {
      try {
        el.setSelectionRange(el.value.length, el.value.length);
      } catch {
        // Some input types (email, number) have no selection.
      }
    } else {
      el.select();
    }
    const expected = append ? el.value + text : text;
    const inserted = text === "" ? document.execCommand("delete") : document.execCommand("insertText", false, text);
    if (!inserted || el.value !== expected) {
      // execCommand failed or was rewritten; set the value the way frameworks notice.
      const setter = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(el), "value")?.set;
      if (setter) setter.call(el, expected);
      else el.value = expected;
      el.dispatchEvent(new Event("input", { bubbles: true }));
    }
    el.dispatchEvent(new Event("change", { bubbles: true }));
    return { ok: true, value: el.type === "password" ? "(hidden)" : el.value };
  }
  if (el.isContentEditable) {
    const selection = getSelection();
    const range = document.createRange();
    range.selectNodeContents(el);
    if (append) range.collapse(false);
    selection?.removeAllRanges();
    selection?.addRange(range);
    if (text === "") document.execCommand("delete");
    else document.execCommand("insertText", false, text);
    return { ok: true, value: el.innerText.slice(0, 200) };
  }
  return { error: `[${ref ?? selector}] is a ${el.tagName.toLowerCase()}, not a text field.` };
}

/** Chooses an option of a <select> by value or label. */
export function selectOption(ref: number | null, selector: string | null, value: string): PageResult<{ selected: string }> {
  const el = (ref ? document.querySelector(`[data-gocode-ref="${ref}"]`) : selector ? document.querySelector(selector) : null) as HTMLElement | null;
  if (!el) return { error: ref ? `There's no element [${ref}] now; take a new snapshot.` : `No element matches ${selector}.` };
  if (!(el instanceof HTMLSelectElement)) return { error: "That isn't a <select>; click it to open it, then click the option." };
  const wanted = value.trim().toLowerCase();
  const option = Array.from(el.options).find((o) => o.value === value || o.text.trim().toLowerCase() === wanted);
  if (!option) return { error: `No option "${value}". Options: ${Array.from(el.options).map((o) => o.text.trim()).join(" | ")}` };
  el.value = option.value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  el.dispatchEvent(new Event("change", { bubbles: true }));
  return { ok: true, selected: option.text.trim() };
}

/** Sends a key press to an element or the focused one. */
export function pressKey(ref: number | null, selector: string | null, key: string): PageResult<{ target: string }> {
  const found = ref ? document.querySelector(`[data-gocode-ref="${ref}"]`) : selector ? document.querySelector(selector) : null;
  if ((ref || selector) && !found) return { error: ref ? `There's no element [${ref}] now; take a new snapshot.` : `No element matches ${selector}.` };
  const target = (found ?? document.activeElement ?? document.body) as HTMLElement;
  if (found) target.focus();
  const codes: Record<string, string> = { " ": "Space", Enter: "Enter", Escape: "Escape", Tab: "Tab", Backspace: "Backspace" };
  const code = codes[key] ?? (key.length === 1 ? (/[a-z]/i.test(key) ? `Key${key.toUpperCase()}` : /\d/.test(key) ? `Digit${key}` : "") : key);
  const init: KeyboardEventInit = { key, code, bubbles: true, cancelable: true, composed: true };
  const proceed = target.dispatchEvent(new KeyboardEvent("keydown", init));
  if (key.length === 1 || key === "Enter") target.dispatchEvent(new KeyboardEvent("keypress", init));
  target.dispatchEvent(new KeyboardEvent("keyup", init));
  if (proceed && key === "Enter") {
    // Synthetic key events don't trigger the browser's own Enter behavior.
    if (target instanceof HTMLInputElement && target.form) target.form.requestSubmit();
    else if (target instanceof HTMLButtonElement || target instanceof HTMLAnchorElement) target.click();
  }
  return { ok: true, target: target.tagName.toLowerCase() };
}

/** Scrolls the page, or an element into view. */
export function scrollPage(ref: number | null, selector: string | null, direction: string): PageResult<{ position: string }> {
  const el = ref ? document.querySelector(`[data-gocode-ref="${ref}"]`) : selector ? document.querySelector(selector) : null;
  if ((ref || selector) && !el) return { error: ref ? `There's no element [${ref}] now; take a new snapshot.` : `No element matches ${selector}.` };
  if (el) {
    el.scrollIntoView({ block: "center" });
  } else {
    // Pages that scroll an inner container instead of the document.
    const root = document.scrollingElement ?? document.documentElement;
    let scroller: Element = root;
    if (root.scrollHeight <= innerHeight + 1) {
      let best = 0;
      for (const candidate of document.querySelectorAll("*")) {
        const overflow = getComputedStyle(candidate).overflowY;
        const area = candidate.clientWidth * candidate.clientHeight;
        if ((overflow === "auto" || overflow === "scroll") && candidate.scrollHeight > candidate.clientHeight + 20 && area > best) {
          best = area;
          scroller = candidate;
        }
      }
    }
    const page = Math.round(scroller === root ? innerHeight * 0.8 : scroller.clientHeight * 0.8);
    const tops: Record<string, number> = {
      up: scroller.scrollTop - page, down: scroller.scrollTop + page, top: 0, bottom: scroller.scrollHeight,
    };
    if (!(direction in tops)) return { error: `Scroll direction must be up, down, top or bottom, not "${direction}".` };
    scroller.scrollTo({ top: tops[direction], behavior: "instant" });
  }
  const root = document.scrollingElement ?? document.documentElement;
  return { ok: true, position: `${Math.round(scrollY)} of ${root.scrollHeight}px (window ${innerHeight}px)` };
}

/** Waits until text appears on the page. */
export async function waitForText(text: string, timeoutMs: number): Promise<PageResult<{ found: boolean }>> {
  const deadline = Date.now() + timeoutMs;
  const wanted = text.toLowerCase();
  while (Date.now() < deadline) {
    if ((document.body?.innerText ?? "").toLowerCase().includes(wanted)) return { ok: true, found: true };
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  return { ok: true, found: false };
}

/** Runs the body of an async function in the page's own JavaScript. */
export async function evaluateScript(script: string): Promise<PageResult<{ value: string }>> {
  let fn: () => Promise<unknown>;
  try {
    fn = new Function(`return (async () => {\n${script}\n})();`) as () => Promise<unknown>;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { error: /unsafe-eval|Content Security Policy/i.test(message) ? `This page's security policy blocks evaluate (${message}). Use snapshot, click and type instead.` : message };
  }
  try {
    const value = await fn();
    if (value === undefined) return { ok: true, value: "undefined" };
    if (value instanceof Element) return { ok: true, value: value.outerHTML.slice(0, 5000) };
    try {
      return { ok: true, value: JSON.stringify(value, null, 2) ?? String(value) };
    } catch {
      return { ok: true, value: String(value) };
    }
  } catch (error) {
    return { error: error instanceof Error ? `${error.name}: ${error.message}` : String(error) };
  }
}
