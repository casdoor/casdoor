// Vite writes to build-temp so that a failed build never leaves a half-written
// "build" directory behind; this swaps it into place.
import fs from "fs";
import path from "path";
import {fileURLToPath} from "url";

const dirname = path.dirname(fileURLToPath(import.meta.url));
const sourceDir = path.join(dirname, "build-temp");
const targetDir = path.join(dirname, "build");
const backupDir = path.join(dirname, "build-old");

// On Windows a removed directory lingers while another process holds a handle
// under it, so let rmSync retry instead of failing the build.
const rmOptions = {recursive: true, force: true, maxRetries: 10, retryDelay: 200};

if (!fs.existsSync(sourceDir)) {
  // eslint-disable-next-line no-console
  console.error(`Source directory "${sourceDir}" does not exist.`);
  process.exit(1);
}

// Left behind by an earlier run that was interrupted between the two renames.
if (fs.existsSync(backupDir)) {
  if (fs.existsSync(targetDir)) {
    fs.rmSync(backupDir, rmOptions);
  } else {
    fs.renameSync(backupDir, targetDir);
  }
}

const hasPreviousBuild = fs.existsSync(targetDir);
if (hasPreviousBuild) {
  fs.renameSync(targetDir, backupDir);
}

try {
  fs.renameSync(sourceDir, targetDir);
} catch (err) {
  if (hasPreviousBuild) {
    fs.renameSync(backupDir, targetDir);
  }
  throw err;
}

if (hasPreviousBuild) {
  fs.rmSync(backupDir, rmOptions);
}

// eslint-disable-next-line no-console
console.log(`Renamed "${sourceDir}" to "${targetDir}" successfully.`);
