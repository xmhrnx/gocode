// Builds the extension into dist/: compiles src/ with tsc and copies
// static/ (manifest, popup page, icons) next to it. Load dist/ as an
// unpacked extension.
import { cpSync, rmSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
rmSync("dist", { recursive: true, force: true });
execFileSync(process.execPath, [require.resolve("typescript/bin/tsc")], { stdio: "inherit" });
cpSync("static", "dist", { recursive: true });
console.log("Built dist/. Load it in chrome://extensions with \"Load unpacked\".");
