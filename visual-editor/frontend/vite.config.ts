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
      // The Examples menu bundles ../../examples/*.layer, and the Help dialog
      // shows the blend-mode animations from ../../docs/images/blend.
      allow: ['.', '../../examples', '../../docs/images'],
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
