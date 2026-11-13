import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  // Relative asset URLs: the built editor works from any sub-path
  // (GitHub Pages serves it at /<repo>/).
  base: './',
  server: {
    open: true,
    fs: {
      // The Examples menu bundles ../../examples/*.layer from the repo root.
      allow: ['.', '../../examples'],
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
