import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [tailwindcss(), svelte()],
  server: {
    proxy: {
      // Proxy /media requests to Wails backend during dev mode
      // This ensures Vite doesn't intercept these requests
      "/media": {
        target: "http://localhost:34115", // Wails dev server default port
        changeOrigin: true,
        ws: false, // WebSocket not needed for media files
      },
    },
  },
});
