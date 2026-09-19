"use strict";

const path = require("node:path");
const { spawnSync } = require("node:child_process");

function managerFromUserAgent(userAgent = "") {
  if (userAgent.startsWith("pnpm/")) return "pnpm";
  if (userAgent.startsWith("yarn/")) return "yarn";
  if (userAgent.startsWith("npm/")) return "npm";
  return null;
}

function updateCommand(manager, packageName) {
  switch (manager) {
    case "npm": return { command: "npm", args: ["install", "--global", `${packageName}@latest`] };
    case "pnpm": return { command: "pnpm", args: ["add", "--global", `${packageName}@latest`] };
    case "yarn": return { command: "yarn", args: ["global", "add", `${packageName}@latest`] };
    default: throw new Error("mdo cannot determine which package manager installed it. Reinstall with npm, pnpm, or Yarn.");
  }
}

function managerFromGlobalRoot(command, args, rootTransform = (root) => root) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  if (result.error || result.status !== 0) return null;
  const globalRoot = rootTransform(result.stdout.trim());
  const packageDirectory = path.resolve(__dirname, "..");
  return packageDirectory.startsWith(`${path.resolve(globalRoot)}${path.sep}`) ? command : null;
}

function installedManager() {
  try {
    const candidates = [
      managerFromGlobalRoot("npm", ["root", "--global"]),
      managerFromGlobalRoot("pnpm", ["root", "--global"]),
      managerFromGlobalRoot("yarn", ["global", "dir"], (directory) => path.join(directory, "node_modules"))
    ].filter(Boolean);
    return candidates.length === 1 ? candidates[0] : null;
  } catch { return null; }
}

function runUpdate(packageName, manager) {
  const specification = updateCommand(manager || installedManager(), packageName);
  const result = spawnSync(specification.command, specification.args, { stdio: "inherit" });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exitCode = result.status || 1;
}

async function checkForUpdate(packageName, installedVersion) {
  const response = await fetch(`https://registry.npmjs.org/${encodeURIComponent(packageName)}/latest`);
  if (!response.ok) throw new Error(`npm registry returned ${response.status} while checking for updates`);
  const latest = (await response.json()).version;
  if (typeof latest !== "string" || !latest) throw new Error("npm returned no latest version for mdo");
  process.stdout.write(latest === installedVersion
    ? `mdo ${installedVersion} is up to date.\n`
    : `mdo ${installedVersion} is installed; ${latest} is available. Run: mdo update\n`);
}

module.exports = { checkForUpdate, installedManager, managerFromGlobalRoot, managerFromUserAgent, runUpdate, updateCommand };
