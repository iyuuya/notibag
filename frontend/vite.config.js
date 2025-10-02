import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      // devOptions: { enabled: true },
      registerType: 'autoUpdate',
      manifest: {
        name: "Notibag",
        short_name: "Notibag",
        description: "Notification bag",
        theme_color: "#ffffff",
        background_color: "#000000",
        start_url: "http://172.16.200.4:8888/",
        display: "standalone",
        icons: [
          {
            src: 'assets/icon-192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'assets/icon-192.png',
            sizes: '512x512',
            type: 'image/png'
          },
        ],
      }
    }),
  ],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist",
    assetsDir: "assets",
  },
});
