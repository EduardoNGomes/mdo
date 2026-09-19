#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");

const [version, artifactsDirectory, outputDirectory] = process.argv.slice(2);
if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/.test(version || "")) {
  throw new Error("usage: prepare-release.js <semver-version> <artifacts-directory> <output-directory>");
}

const sourceDirectory = path.resolve(__dirname, "..");
const artifacts = path.resolve(artifactsDirectory);
const output = path.resolve(outputDirectory);
const targets = [
  ["darwin", "arm64"], ["darwin", "x64"],
  ["linux", "arm64"], ["linux", "x64"],
  ["win32", "arm64"], ["win32", "x64"]
];

fs.rmSync(output, { recursive: true, force: true });
fs.mkdirSync(output, { recursive: true });
for (const directory of ["bin", "lib"]) {
  fs.cpSync(path.join(sourceDirectory, directory), path.join(output, "mdo", directory), { recursive: true });
}

const rootManifest = JSON.parse(fs.readFileSync(path.join(sourceDirectory, "package.json"), "utf8"));
rootManifest.version = version;
for (const packageName of Object.keys(rootManifest.optionalDependencies)) {
  rootManifest.optionalDependencies[packageName] = version;
}
fs.writeFileSync(path.join(output, "mdo", "package.json"), `${JSON.stringify(rootManifest, null, 2)}\n`);

for (const [platform, arch] of targets) {
  const target = `${platform}-${arch}`;
  const packageName = `@egomes.dev/mdo-${target}`;
  const binaryName = platform === "win32" ? "mdo.exe" : "mdo";
  const sourceBinary = path.join(artifacts, `mdo-${target}${platform === "win32" ? ".exe" : ""}`);
  if (!fs.existsSync(sourceBinary)) throw new Error(`missing release binary: ${sourceBinary}`);

  const packageDirectory = path.join(output, `mdo-${target}`);
  fs.mkdirSync(packageDirectory, { recursive: true });
  fs.copyFileSync(sourceBinary, path.join(packageDirectory, binaryName));
  if (platform !== "win32") fs.chmodSync(path.join(packageDirectory, binaryName), 0o755);
  fs.writeFileSync(path.join(packageDirectory, "package.json"), `${JSON.stringify({
    name: packageName,
    version,
    description: `Native ${target} binary for mdo`,
    repository: rootManifest.repository,
    os: [platform],
    cpu: [arch],
    files: [binaryName]
  }, null, 2)}\n`);
}
