const fs = require("fs");
const path = require("path");

const sourceDir = path.join(__dirname, "build-temp");
const targetDir = path.join(__dirname, "build");
const backupDir = path.join(__dirname, "build-old");

// On Windows a directory can linger in a "delete pending" state while another
// process still holds a handle to a file inside it: the Casdoor server serves
// files straight out of build/, and virus scanners open freshly written ones.
// Removes and renames then fail with EPERM/EBUSY, so both are retried here.
const RETRY_ATTEMPTS = 10;
const RETRY_DELAY_MS = 200;
const RETRYABLE_CODES = ["EPERM", "EBUSY", "EACCES", "ENOTEMPTY"];

const rmOptions = {recursive: true, force: true, maxRetries: RETRY_ATTEMPTS, retryDelay: RETRY_DELAY_MS};

function sleepSync(ms) {
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, ms);
}

function renameSyncWithRetry(from, to) {
  for (let attempt = 0; ; attempt++) {
    try {
      fs.renameSync(from, to);
      return;
    } catch (err) {
      if (attempt >= RETRY_ATTEMPTS || !RETRYABLE_CODES.includes(err.code)) {
        throw err;
      }
      sleepSync(RETRY_DELAY_MS);
    }
  }
}

if (!fs.existsSync(sourceDir)) {
  // eslint-disable-next-line no-console
  console.error(`Source directory "${sourceDir}" does not exist.`);
  process.exit(1);
}

// Left behind by an earlier run that was interrupted between the two renames.
if (fs.existsSync(backupDir)) {
  fs.rmSync(backupDir, rmOptions);
}

// Move the previous build aside rather than deleting it up front. If anything
// below fails, build/ is either still there or gets restored, instead of the
// server being left with no build to serve at all.
const hasPreviousBuild = fs.existsSync(targetDir);
if (hasPreviousBuild) {
  renameSyncWithRetry(targetDir, backupDir);
}

try {
  renameSyncWithRetry(sourceDir, targetDir);
} catch (err) {
  if (hasPreviousBuild) {
    renameSyncWithRetry(backupDir, targetDir);
    // eslint-disable-next-line no-console
    console.error(`Could not move "${sourceDir}" into place, restored the previous build.`);
  }
  throw err;
}

if (hasPreviousBuild) {
  fs.rmSync(backupDir, rmOptions);
}

// eslint-disable-next-line no-console
console.log(`Renamed "${sourceDir}" to "${targetDir}" successfully.`);
