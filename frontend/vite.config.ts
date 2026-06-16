import { defineConfig } from "vite";
import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  plugins: [tailwindcss(), sveltekit(), wails("./bindings")],
});
