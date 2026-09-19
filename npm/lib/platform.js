"use strict";

const targets = {
  "darwin-arm64": { packageName: "@egomes.dev/mdo-darwin-arm64", binary: "mdo" },
  "darwin-x64": { packageName: "@egomes.dev/mdo-darwin-x64", binary: "mdo" },
  "linux-arm64": { packageName: "@egomes.dev/mdo-linux-arm64", binary: "mdo" },
  "linux-x64": { packageName: "@egomes.dev/mdo-linux-x64", binary: "mdo" },
  "win32-arm64": { packageName: "@egomes.dev/mdo-win32-arm64", binary: "mdo.exe" },
  "win32-x64": { packageName: "@egomes.dev/mdo-win32-x64", binary: "mdo.exe" }
};

function targetFor(platform = process.platform, arch = process.arch) {
  const target = targets[`${platform}-${arch}`];
  if (!target) {
    throw new Error(
      `mdo does not support ${platform}/${arch}. Supported targets: ${Object.keys(targets).join(", ")}.`
    );
  }
  return target;
}

module.exports = { targetFor };
