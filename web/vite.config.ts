import path from "path";
import {defineConfig} from "vite";
import react from "@vitejs/plugin-react";

const backend = process.env.CASDOOR_BACKEND || "http://localhost:8000";

const proxyPaths = [
  "/api",
  "/swagger",
  "/files",
  "/.well-known/openid-configuration",
  "/scim",
  "/cas",
];

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 7001,
    proxy: Object.fromEntries(
      proxyPaths.map((p) => [p, {target: backend, changeOrigin: true}])
    ),
  },
  build: {
    outDir: "build-temp",
    sourcemap: false,
    chunkSizeWarningLimit: 2000,
  },
  experimental: {
    // index.html keeps "/assets/..." (the server may rewrite it to a CDN), while the chunks and assets that js and css
    // load are resolved relative to the file loading them, so they come from the same place as the main bundle
    renderBuiltUrl(filename, {hostType}) {
      if (hostType === "js" || hostType === "css") {
        return {relative: true};
      }
      return {relative: false};
    },
  },
});
