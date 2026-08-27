// Vite writes to build-temp so that a failed build never leaves a half-written
// "build" directory behind; this renames it into place, like web/mv.js does.
import fs from "fs";
import path from "path";
import {fileURLToPath} from "url";

const dirname = path.dirname(fileURLToPath(import.meta.url));
const sourceDir = path.join(dirname, "build-temp");
const targetDir = path.join(dirname, "build");

if (!fs.existsSync(sourceDir)) {
  // eslint-disable-next-line no-console
  console.error(`Source directory "${sourceDir}" does not exist.`);
  process.exit(1);
}

if (fs.existsSync(targetDir)) {
  fs.rmSync(targetDir, {recursive: true, force: true});
  // eslint-disable-next-line no-console
  console.log(`Target directory "${targetDir}" has been deleted successfully.`);
}

fs.renameSync(sourceDir, targetDir);
// eslint-disable-next-line no-console
console.log(`Renamed "${sourceDir}" to "${targetDir}" successfully.`);
