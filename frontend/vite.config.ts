import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// `vite --mode demo` swaps the generated Wails service for an in-memory mock (screenshots only).
export default defineConfig(({ mode }) => ({
  plugins: [vue()],
  resolve: mode === 'demo'
    ? { alias: [{ find: /^.*\/bindings\/github\.com\/WilliamFalci\/capsuleers\.settings\/service$/, replacement: fileURLToPath(new URL('./src/demo/service.ts', import.meta.url)) }] }
    : {},
  build: { outDir: 'dist', emptyOutDir: true, target: 'es2022', assetsInlineLimit: 0 },
}))
