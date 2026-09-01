// The console is an ESM package ("type": "module"), so this config uses
// `export default` rather than the `module.exports` the antd frontend used.
import {defineConfig} from "cypress";

export default defineConfig({
  e2e: {
    baseUrl: "http://localhost:7001",
    retries: {
      runMode: 2,
      openMode: 0,
    },
    // the console is a dense admin UI; a wider viewport keeps the table toolbar
    // on one row so the Add button is not pushed out of view
    viewportWidth: 1400,
    viewportHeight: 900,
  },
});
