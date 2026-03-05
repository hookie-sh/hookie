/**
 * Postinstall: download the Hookie CLI binary for this platform from GitHub releases.
 * Reads version from package.json and fetches hookie_${version}_${os}_${arch}.tar.gz or .zip.
 */

const fs = require("fs");
const path = require("path");
const https = require("https");

const pkgPath = path.join(__dirname, "..", "package.json");
const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
const version = pkg.version;
if (!version || version === "0.0.0") {
  console.warn("@hookie/cli: skipping postinstall (version 0.0.0 or not set)");
  process.exit(0);
}

const tag = `v${version}`;
const platform = process.platform;
const arch = process.arch;

// Map Node to GoReleaser os/arch
const osMap = { darwin: "darwin", linux: "linux", win32: "windows" };
const archMap = { x64: "amd64", arm64: "arm64" };
const goOs = osMap[platform];
const goArch = archMap[arch];
if (!goOs || !goArch) {
  console.error(`@hookie/cli: unsupported platform ${platform}/${arch}`);
  process.exit(1);
}

const ext = platform === "win32" ? "zip" : "tar.gz";
const artifactName = `hookie_${version}_${goOs}_${goArch}.${ext}`;
const url = `https://github.com/hookie-sh/hookie/releases/download/${tag}/${artifactName}`;

const binDir = path.join(__dirname, "..", "bin");
const binaryName = platform === "win32" ? "hookie.exe" : "hookie";
const binaryPath = path.join(binDir, binaryName);

if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

function download(url, outputFileName = artifactName) {
  return new Promise((resolve, reject) => {
    const file = path.join(binDir, outputFileName);
    const stream = fs.createWriteStream(file);
    https
      .get(url, { redirect: true }, (res) => {
        if (res.statusCode === 302 || res.statusCode === 301) {
          download(res.headers.location, outputFileName).then(resolve).catch(reject);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode} for ${url}`));
          return;
        }
        res.pipe(stream);
        stream.on("finish", () => {
          stream.close();
          resolve(file);
        });
      })
      .on("error", reject);
  });
}

async function main() {
  try {
    const archivePath = await download(url);
    function ensureBinaryInPlace() {
      if (fs.existsSync(binaryPath)) return;
      const names = fs.readdirSync(binDir);
      const subdir = names.find((n) => n.startsWith("hookie_"));
      if (subdir) {
        const src = path.join(binDir, subdir, binaryName);
        if (fs.existsSync(src)) {
          fs.renameSync(src, binaryPath);
          fs.rmSync(path.join(binDir, subdir), { recursive: true });
        }
      }
    }

    if (ext === "zip") {
      const AdmZip = require("adm-zip");
      const zip = new AdmZip(archivePath);
      const entries = zip.getEntries();
      const exeEntry = entries.find((e) => !e.isDirectory && (e.entryName === binaryName || e.entryName.endsWith(`/${binaryName}`)));
      if (!exeEntry) {
        throw new Error(`No ${binaryName} in archive`);
      }
      zip.extractAllTo(binDir, true);
      ensureBinaryInPlace();
    } else {
      const tar = require("tar");
      await tar.extract({ file: archivePath, cwd: binDir });
      ensureBinaryInPlace();
    }
    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary not found at ${binaryPath} after extract`);
    }
    fs.unlinkSync(archivePath);
    if (platform !== "win32") {
      fs.chmodSync(binaryPath, 0o755);
    }
  } catch (err) {
    console.error("@hookie/cli postinstall failed:", err.message);
    process.exit(1);
  }
}

main();
