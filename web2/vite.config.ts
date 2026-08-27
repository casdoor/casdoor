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
    port: 7002,
    proxy: Object.fromEntries(
      proxyPaths.map((p) => [p, {target: backend, changeOrigin: true}])
    ),
  },
  build: {
    outDir: "build-temp",
    sourcemap: false,
    chunkSizeWarningLimit: 2000,
  },
});
