import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      pages: 'dist',
      assets: 'dist',
      fallback: 'index.html',
      precompress: false,
      strict: false,
    }),
    alias: {
      '@domain': 'src/domain',
      '@application': 'src/application/index.ts',
      '@application/*': 'src/application/*',
      '@infrastructure/wails': 'src/infrastructure/wails/index.ts',
      '@infrastructure': 'src/infrastructure',
      '@infrastructure/*': 'src/infrastructure/*',
      '@presentation': 'src/presentation',
      '@presentation/*': 'src/presentation/*',
      '@bindings': 'bindings',
      '@bindings/*': 'bindings/*',
    },
  },
};

export default config;
