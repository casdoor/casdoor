// web2 is an ESM package ("type": "module"), so this config uses `export default`
// rather than the `module.exports` the antd frontend's config uses.
import {defineConfig} from "cypress";

export default defineConfig({
  e2e: {
    baseUrl: "http://localhost:7002",
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
