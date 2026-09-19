"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");
const { targetFor } = require("../lib/platform");

test("selects the Linux x64 package", () => {
  assert.deepEqual(targetFor("linux", "x64"), {
    packageName: "@egomes.dev/mdo-linux-x64",
    binary: "mdo"
  });
});

test("rejects unsupported targets", () => {
  assert.throws(() => targetFor("freebsd", "x64"), /does not support freebsd\/x64/);
});
