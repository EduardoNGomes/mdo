"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");
const { managerFromUserAgent, updateCommand } = require("../lib/update");

test("recognizes package-manager user agents", () => {
  assert.equal(managerFromUserAgent("npm/11.0.0 node/v24"), "npm");
  assert.equal(managerFromUserAgent("pnpm/10.0.0 npm/? node/v24"), "pnpm");
  assert.equal(managerFromUserAgent("yarn/1.22.0 npm/? node/v24"), "yarn");
});

test("builds a fixed npm update command", () => {
  assert.deepEqual(updateCommand("npm", "mdo"), {
    command: "npm",
    args: ["install", "--global", "mdo@latest"]
  });
});
