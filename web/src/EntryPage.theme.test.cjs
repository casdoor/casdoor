"use strict";

const assert = require("assert");
const fs = require("fs");
const path = require("path");

const src = fs.readFileSync(path.join(__dirname, "EntryPage.js"), "utf8");

const applicationUpdate = src.match(
  /const onUpdateApplication = \(application\) => \{[\s\S]*?this\.props\.updataThemeData\(([^)]+)\)/
);
assert.ok(applicationUpdate, "onUpdateApplication should call updataThemeData");
assert.equal(
  applicationUpdate[1].trim(),
  "themeData",
  "loading an application on login/entry must not pass InitThemeAlgorithm; that re-applies a stored dark algorithm and stuck the UI in dark mode after dbfe1c6 (casdoor#5771)"
);

console.log("ok - EntryPage application theme update does not re-init algorithm");
