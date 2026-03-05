#!/usr/bin/env node
/**
 * Hookie CLI wrapper: resolves the platform-specific binary from optionalDependencies
 * and spawns it. Requires Node 18+.
 */

const { spawnSync } = require("child_process");
const path = require("path");

// Map process.platform / process.arch to npm package name (matches optionalDependencies)
const PLATFORM_PACKAGES = {
  "darwin arm64": "@hookie-sh/hookie-darwin-arm64",
  "darwin x64": "@hookie-sh/hookie-darwin-x64",
  "linux arm64": "@hookie-sh/hookie-linux-arm64",
  "linux x64": "@hookie-sh/hookie-linux-x64",
  "win32 x64": "@hookie-sh/hookie-win32-x64",
};

const platformKey = `${process.platform} ${process.arch}`;
const pkgName = PLATFORM_PACKAGES[platformKey];

if (!pkgName) {
  console.error(
    `hookie: unsupported platform ${process.platform}/${process.arch}. ` +
      "Supported: darwin (arm64, x64), linux (arm64, x64), win32 (x64)."
  );
  process.exit(1);
}

const binaryName = process.platform === "win32" ? "hookie.exe" : "hookie";
const subpath = `bin/${binaryName}`;

let binaryPath;
try {
  binaryPath = require.resolve(`${pkgName}/${subpath}`);
} catch (err) {
  console.error(
    `hookie: could not find platform binary (${pkgName}). ` +
      "If you used --no-optional or --omit=optional, install without that flag. " +
      "Otherwise, this platform may not be supported."
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: true,
});

process.exit(result.status ?? result.signal ? 128 + result.signal : 0);
