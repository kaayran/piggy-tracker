import { defineConfig, loadEnv, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

const devProfiles = [
  { id: 1, first_name: "Ann", username: "dev_ann" },
  { id: 2, first_name: "Bob", username: "dev_bob" },
  { id: 3, first_name: "Carl", username: "dev_carl" },
];

const hex = (buf: ArrayBuffer) =>
  [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, "0")).join("");

async function signLogin(botToken: string, fields: Record<string, string>) {
  const encoder = new TextEncoder();
  const secret = await crypto.subtle.digest("SHA-256", encoder.encode(botToken));
  const key = await crypto.subtle.importKey("raw", secret, { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
  const checkString = Object.keys(fields)
    .sort()
    .map((k) => `${k}=${fields[k]}`)
    .join("\n");
  const hash = hex(await crypto.subtle.sign("HMAC", key, encoder.encode(checkString)));
  return new URLSearchParams({ ...fields, hash }).toString();
}

function devLogin(botToken: string): Plugin {
  return {
    name: "dev-login",
    configureServer(server) {
      server.middlewares.use("/dev/profiles", async (req, res) => {
        const host = ((req as { headers: { host?: string } }).headers.host ?? "").split(":")[0];
        if (host !== "localhost" && host !== "127.0.0.1") {
          res.statusCode = 404;
          res.end();
          return;
        }
        if (!botToken) {
          res.statusCode = 500;
          res.end(JSON.stringify({ code: "no_bot_token", message: "BOT_TOKEN is not set in .env" }));
          return;
        }
        const authDate = String(Math.floor(Date.now() / 1000));
        const profiles = await Promise.all(
          devProfiles.map(async (p) => ({
            name: p.first_name,
            payload: await signLogin(botToken, {
              id: String(p.id),
              first_name: p.first_name,
              username: p.username,
              auth_date: authDate,
            }),
          })),
        );
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify(profiles));
      });
    },
  };
}

export default defineConfig(({ mode }) => ({
  envDir: "..",
  plugins: [react(), devLogin(loadEnv(mode, "..", "").BOT_TOKEN ?? "")],
  // The output dir is not emptied: it holds a committed .gitkeep so the Go embed always
  // has something to match. `make build` clears stale assets instead.
  build: { outDir: "../api/webdist", emptyOutDir: false },
  server: {
    allowedHosts: ["dev.piggy.overstats.org"],
    hmr: { clientPort: 443 },
    proxy: { "/api": "http://127.0.0.1:8080" },
  },
}));
