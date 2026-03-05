#!/usr/bin/env node
/**
 * Publish platform-specific npm packages from GoReleaser artifacts.
 * Run from repo root with: node packages/hookie-cli/scripts/publish-platform-packages.js <version> [distDir]
 * Example: node packages/hookie-cli/scripts/publish-platform-packages.js 1.0.0 dist
 * Requires NODE_AUTH_TOKEN or NPM_AUTH_TOKEN for npm publish.
 */

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const os = require("os");

const version = process.argv[2];
const distDir = process.argv[3] || "dist";

if (!version || !/^\d+\.\d+\.\d+/.test(version)) {
  console.error("Usage: node publish-platform-packages.js <version> [distDir]");
  console.error("Example: node publish-platform-packages.js 1.0.0 dist");
  process.exit(1);
}

const distPath = path.resolve(process.cwd(), distDir);
if (!fs.existsSync(distPath)) {
  console.error(`Dist directory not found: ${distPath}`);
  process.exit(1);
}

// Map GoReleaser artifact (goos_goarch) to npm package name and os/cpu
const ARTIFACT_MAP = [
  { artifact: "darwin_amd64", pkg: "@hookie-sh/hookie-darwin-x64", os: ["darwin"], cpu: ["x64"] },
  { artifact: "darwin_arm64", pkg: "@hookie-sh/hookie-darwin-arm64", os: ["darwin"], cpu: ["arm64"] },
  { artifact: "linux_amd64", pkg: "@hookie-sh/hookie-linux-x64", os: ["linux"], cpu: ["x64"] },
  { artifact: "linux_arm64", pkg: "@hookie-sh/hookie-linux-arm64", os: ["linux"], cpu: ["arm64"] },
  { artifact: "windows_amd64", pkg: "@hookie-sh/hookie-win32-x64", os: ["win32"], cpu: ["x64"] },
];

function extractBinary(archivePath, outDir, binaryName) {
  const ext = path.extname(archivePath);
  const binDir = path.join(outDir, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  if (ext === ".zip") {
    const AdmZip = require("adm-zip");
    const zip = new AdmZip(archivePath);
    const entries = zip.getEntries();
    const exeEntry = entries.find(
      (e) => !e.isDirectory && (e.entryName === binaryName || e.entryName.endsWith(`/${binaryName}`))
    );
    if (!exeEntry) {
      throw new Error(`No ${binaryName} in archive`);
    }
    zip.extractAllTo(outDir, true);
    // GoReleaser may put binary in a subdir like hookie_1.0.0_windows_amd64/
    const names = fs.readdirSync(outDir);
    const subdir = names.find((n) => n.startsWith("hookie_"));
    if (subdir) {
      const src = path.join(outDir, subdir, binaryName);
      if (fs.existsSync(src)) {
        fs.renameSync(src, path.join(binDir, binaryName));
        fs.rmSync(path.join(outDir, subdir), { recursive: true });
      }
    } else if (fs.existsSync(path.join(outDir, binaryName))) {
      fs.renameSync(path.join(outDir, binaryName), path.join(binDir, binaryName));
    }
  } else {
    const tar = require("tar");
    tar.extract({ file: archivePath, cwd: outDir, sync: true });
    const names = fs.readdirSync(outDir);
    const subdir = names.find((n) => n.startsWith("hookie_"));
    if (subdir) {
      const src = path.join(outDir, subdir, binaryName);
      if (fs.existsSync(src)) {
        fs.mkdirSync(binDir, { recursive: true });
        fs.renameSync(src, path.join(binDir, binaryName));
        fs.rmSync(path.join(outDir, subdir), { recursive: true });
      }
    } else if (fs.existsSync(path.join(outDir, binaryName))) {
      fs.mkdirSync(binDir, { recursive: true });
      fs.renameSync(path.join(outDir, binaryName), path.join(binDir, binaryName));
    }
  }

  const binaryPath = path.join(binDir, binaryName);
  if (!fs.existsSync(binaryPath)) {
    throw new Error(`Binary not found at ${binaryPath} after extract`);
  }
  if (binaryName !== "hookie.exe") {
    fs.chmodSync(binaryPath, 0o755);
  }
}

function main() {
  for (const { artifact, pkg, os: pkgOs, cpu } of ARTIFACT_MAP) {
    const ext = artifact.startsWith("windows_") ? "zip" : "tar.gz";
    const artifactName = `hookie_${version}_${artifact}.${ext}`;
    const archivePath = path.join(distPath, artifactName);

    if (!fs.existsSync(archivePath)) {
      console.error(`Artifact not found: ${archivePath}`);
      process.exit(1);
    }

    const binaryName = artifact.startsWith("windows_") ? "hookie.exe" : "hookie";
    const workDir = fs.mkdtempSync(path.join(os.tmpdir(), `hookie-pkg-${artifact}-`));

    try {
      extractBinary(archivePath, workDir, binaryName);

      const pkgJson = {
        name: pkg,
        version,
        description: `Hookie CLI binary for ${pkgOs[0]}/${cpu[0]}`,
        repository: { type: "git", url: "https://github.com/hookie-sh/hookie.git" },
        license: "Apache-2.0",
        os: pkgOs,
        cpu,
        engines: { node: ">=18" },
        publishConfig: { access: "public", provenance: true },
      };

      fs.writeFileSync(
        path.join(workDir, "package.json"),
        JSON.stringify(pkgJson, null, 2)
      );

      console.log(`Publishing ${pkg}@${version}...`);
      execSync("npm publish --provenance", {
        cwd: workDir,
        stdio: "inherit",
      });
      console.log(`Published ${pkg}@${version}`);
    } finally {
      fs.rmSync(workDir, { recursive: true, force: true });
    }
  }
}

main();
