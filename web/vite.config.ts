import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  envDir: "..",
  plugins: [react()],
  // The output dir is not emptied: it holds a committed .gitkeep so the Go embed always
  // has something to match. `make build` clears stale assets instead.
  build: { outDir: "../api/webdist", emptyOutDir: false },
  server: {
    allowedHosts: ["dev.piggy.overstats.org"],
    hmr: { clientPort: 443 },
    proxy: { "/api": "http://127.0.0.1:8080" },
  },
});
