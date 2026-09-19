#!/usr/bin/env node
"use strict";

const path = require("node:path");
const { spawnSync } = require("node:child_process");
const { targetFor } = require("../lib/platform");
const { checkForUpdate, runUpdate } = require("../lib/update");
const packageInfo = require("../package.json");

async function main(args = process.argv.slice(2)) {
  if (args[0] === "update") {
    if (args.length === 2 && args[1] === "--check") {
      await checkForUpdate(packageInfo.name, packageInfo.version);
      return;
    }
    let manager;
    if (args.length === 3 && args[1] === "--manager" && ["npm", "pnpm", "yarn"].includes(args[2])) {
      manager = args[2];
    } else if (args.length !== 1) {
      throw new Error("usage: mdo update [--manager npm|pnpm|yarn]");
    }
    runUpdate(packageInfo.name, manager);
    return;
  }

  const target = targetFor();
  const binary = require.resolve(path.posix.join(target.packageName, target.binary));
  const result = spawnSync(binary, args, { stdio: "inherit" });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exitCode = result.status || 1;
}

main().catch((error) => {
  console.error(`mdo: ${error.message}`);
  process.exitCode = 1;
});

module.exports = { main };
