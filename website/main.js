"use strict";

const REPO = "xmhrnx/gocode";
const RELEASES = `https://github.com/${REPO}/releases/latest`;
const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)").matches;

// ---------------------------------------------------------------- background

// A slow field of points that link up near each other and near the pointer.
function startField() {
  const canvas = document.getElementById("field");
  const ctx = canvas.getContext("2d");
  const pointer = { x: -1e4, y: -1e4 };
  let points = [];
  let width = 0, height = 0, ratio = 1;

  function resize() {
    ratio = Math.min(devicePixelRatio || 1, 2);
    width = innerWidth;
    height = innerHeight;
    canvas.width = width * ratio;
    canvas.height = height * ratio;
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    const count = Math.round(Math.min(90, (width * height) / 16000));
    points = Array.from({ length: count }, () => ({
      x: Math.random() * width,
      y: Math.random() * height,
      vx: (Math.random() - 0.5) * 0.25,
      vy: (Math.random() - 0.5) * 0.25,
      r: Math.random() * 1.4 + 0.4,
    }));
  }

  function draw() {
    ctx.clearRect(0, 0, width, height);
    const link = 130;
    for (let i = 0; i < points.length; i++) {
      const p = points[i];
      p.x = (p.x + p.vx + width) % width;
      p.y = (p.y + p.vy + height) % height;
      for (let j = i + 1; j < points.length; j++) {
        const q = points[j];
        const d = Math.hypot(p.x - q.x, p.y - q.y);
        if (d < link) {
          ctx.strokeStyle = `rgba(125,140,255,${0.14 * (1 - d / link)})`;
          ctx.lineWidth = 1;
          ctx.beginPath();
          ctx.moveTo(p.x, p.y);
          ctx.lineTo(q.x, q.y);
          ctx.stroke();
        }
      }
      const dp = Math.hypot(p.x - pointer.x, p.y - pointer.y);
      if (dp < 180) {
        ctx.strokeStyle = `rgba(125,249,196,${0.25 * (1 - dp / 180)})`;
        ctx.beginPath();
        ctx.moveTo(p.x, p.y);
        ctx.lineTo(pointer.x, pointer.y);
        ctx.stroke();
      }
      ctx.fillStyle = "rgba(200,206,255,.55)";
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  let frame = 0;
  function loop() {
    draw();
    frame = requestAnimationFrame(loop);
  }

  resize();
  addEventListener("resize", resize);
  addEventListener("pointermove", (e) => { pointer.x = e.clientX; pointer.y = e.clientY; });
  if (reducedMotion) {
    draw();
    return;
  }
  // Nothing draws while the tab is hidden.
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) cancelAnimationFrame(frame);
    else loop();
  });
  loop();
}

// ---------------------------------------------------------------- hero window

// Types the agent's reply, then shows its tool calls one by one.
function playAgent() {
  const typed = document.querySelector(".typed");
  const tools = [...document.querySelectorAll(".tool")];
  const text = typed.dataset.text;
  const reveal = () => tools.forEach((t, i) => setTimeout(() => t.classList.add("show"), 350 + i * 650));
  if (reducedMotion) {
    typed.textContent = text;
    typed.classList.add("done");
    reveal();
    return;
  }
  let i = 0;
  const tick = () => {
    typed.textContent = text.slice(0, ++i);
    if (i < text.length) setTimeout(tick, 18 + Math.random() * 30);
    else { typed.classList.add("done"); reveal(); }
  };
  setTimeout(tick, 700);
}

// ---------------------------------------------------------------- scroll effects

function countUp(el) {
  const target = Number(el.dataset.count);
  if (reducedMotion || target === 0) { el.textContent = String(target); return; }
  const start = performance.now();
  const step = (now) => {
    const t = Math.min(1, (now - start) / 1400);
    el.textContent = String(Math.round(target * (1 - Math.pow(1 - t, 3))));
    if (t < 1) requestAnimationFrame(step);
  };
  requestAnimationFrame(step);
}

function watchReveals() {
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (!entry.isIntersecting) continue;
      entry.target.classList.add("in");
      entry.target.querySelectorAll("[data-count]").forEach(countUp);
      observer.unobserve(entry.target);
    }
  }, { threshold: 0.25 });
  document.querySelectorAll(".reveal").forEach((el) => observer.observe(el));
  document.querySelectorAll(".hero [data-count]").forEach((el) => setTimeout(() => countUp(el), 500));
}

// Cards light up where the pointer is.
function trackCardGlow() {
  document.querySelectorAll(".card").forEach((card) => {
    card.addEventListener("pointermove", (e) => {
      const box = card.getBoundingClientRect();
      card.style.setProperty("--mx", `${e.clientX - box.left}px`);
      card.style.setProperty("--my", `${e.clientY - box.top}px`);
    });
  });
}

// ---------------------------------------------------------------- download

// detectPlatform reports the visitor's OS and CPU ("x64" or "arm64").
async function detectPlatform() {
  const ua = navigator.userAgent;
  const platform = (navigator.userAgentData && navigator.userAgentData.platform) || navigator.platform || "";
  let os = "unknown";
  if (/win/i.test(platform) || /Windows/.test(ua)) os = "windows";
  else if (/mac/i.test(platform) || /Mac OS X/.test(ua)) os = "macos";
  else if (/linux/i.test(platform) || /Linux|X11/.test(ua)) os = "linux";
  if (/Android|iPhone|iPad/.test(ua)) os = "mobile";

  let arch = /aarch64|arm64|armv8/i.test(ua + platform) ? "arm64" : "x64";
  if (navigator.userAgentData && navigator.userAgentData.getHighEntropyValues) {
    try {
      const v = await navigator.userAgentData.getHighEntropyValues(["architecture", "bitness"]);
      if (v.architecture) arch = v.architecture === "arm" ? "arm64" : "x64";
    } catch { /* keep the guess from the user agent */ }
  }
  return { os, arch };
}

const isArm = (name) => /arm64|aarch64/i.test(name);

// describeAsset names a release file for people: its OS, CPU and kind.
function describeAsset(name) {
  const cpu = /universal/i.test(name) ? "Universal" : isArm(name) ? "ARM64" : "x64";
  if (/\.exe$/i.test(name)) return { os: "windows", label: `Windows ${cpu} installer` };
  if (/\.dmg$/i.test(name)) return { os: "macos", label: `macOS ${cpu === "x64" ? "Intel" : cpu === "ARM64" ? "Apple silicon" : "Universal"}` };
  if (/\.deb$/i.test(name)) return { os: "linux", label: `Linux ${cpu} .deb` };
  if (/\.rpm$/i.test(name)) return { os: "linux", label: `Linux ${cpu} .rpm` };
  if (/\.AppImage$/i.test(name)) return { os: "linux", label: `Linux ${cpu} AppImage` };
  return null;
}

// pickAsset chooses the best file for the platform from the release.
function pickAsset(assets, { os, arch }) {
  const wantArm = arch === "arm64";
  const sameCpu = (a) => /universal/i.test(a.name) || isArm(a.name) === wantArm;
  const preferences = {
    windows: [/setup\.exe$/i, /\.exe$/i],
    macos: [/universal\.dmg$/i, /\.dmg$/i],
    linux: [/\.deb$/i, /\.AppImage$/i, /\.rpm$/i],
  }[os] || [];
  for (const pattern of preferences) {
    const match = assets.find((a) => pattern.test(a.name) && sameCpu(a)) || (os === "macos" && assets.find((a) => pattern.test(a.name)));
    if (match) return match;
  }
  return null;
}

const osNames = { windows: "Windows", macos: "macOS", linux: "Linux" };
const megabytes = (bytes) => `${Math.max(1, Math.round(bytes / 1048576))} MB`;

async function setupDownload() {
  const labels = document.querySelectorAll("[data-download-label]");
  const metas = document.querySelectorAll("[data-download-meta]");
  const buttons = document.querySelectorAll("[data-download]");
  const detected = document.querySelector("[data-download-detected]");
  const list = document.querySelector("[data-download-list]");
  const target = await detectPlatform();
  const cpuName = target.os === "macos" ? "" : target.arch === "arm64" ? " (ARM64)" : " (x64)";
  const osName = osNames[target.os];

  if (osName) {
    labels.forEach((l) => { l.textContent = `Download for ${osName}${cpuName}`; });
    detected.textContent = `Detected ${osName}${cpuName}. Free and open source.`;
  }

  let release;
  try {
    const response = await fetch(`https://api.github.com/repos/${REPO}/releases/latest`, { headers: { Accept: "application/vnd.github+json" } });
    if (!response.ok) throw new Error(String(response.status));
    release = await response.json();
  } catch {
    metas.forEach((m) => { m.textContent = "Downloads are on GitHub Releases."; });
    return;
  }

  const assets = release.assets || [];
  const best = pickAsset(assets, target);
  if (best) {
    buttons.forEach((b) => { b.href = best.browser_download_url; });
    metas.forEach((m) => { m.textContent = `${release.tag_name} · ${megabytes(best.size)} · ${best.name}`; });
  } else {
    buttons.forEach((b) => { b.href = release.html_url || RELEASES; });
    metas.forEach((m) => { m.textContent = `${release.tag_name} · pick your platform below`; });
    if (osName) labels.forEach((l) => { l.textContent = "See all downloads"; });
  }

  const items = assets
    .map((a) => ({ asset: a, info: describeAsset(a.name) }))
    .filter((x) => x.info)
    .sort((x, y) => x.info.label.localeCompare(y.info.label));
  if (items.length === 0) return;
  list.replaceChildren(...items.map(({ asset, info }) => {
    const li = document.createElement("li");
    const a = document.createElement("a");
    a.href = asset.browser_download_url;
    const name = document.createElement("span");
    name.textContent = info.label;
    const size = document.createElement("small");
    size.textContent = megabytes(asset.size);
    a.append(name, size);
    li.append(a);
    return li;
  }));
}

startField();
playAgent();
watchReveals();
trackCardGlow();
setupDownload();
