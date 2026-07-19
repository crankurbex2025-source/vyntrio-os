import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const publicHost = process.env.VITE_DEV_PUBLIC_HOST;
const publicPort = Number(process.env.VITE_DEV_PUBLIC_PORT ?? 5174);

export default defineConfig({
  plugins: [react()],
  server: {
    host: process.env.VITE_DEV_HOST ?? "127.0.0.1",
    port: Number(process.env.VITE_DEV_PORT ?? 5173),
    strictPort: true,
    allowedHosts: ["vyntrio.xyz", "localhost", ".vyntrio.xyz"],
    ...(publicHost
      ? {
          hmr: {
            protocol: "wss",
            host: publicHost,
            clientPort: publicPort,
          },
        }
      : {}),
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules/gsap") || id.includes("node_modules/@gsap")) {
            return "preview-gsap";
          }
        },
      },
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setupMatchMedia.ts"],
  },
});
