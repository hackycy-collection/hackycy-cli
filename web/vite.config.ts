import { resolve } from 'node:path'
import process from 'node:process'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

const apps = {
  'diff': { port: 5173, proxy: ['/api', '/mcp'] },
  'fs': { port: 5174, proxy: ['/api', '/files', '/thumbnails'] },
  'tunnel-server': { port: 5175, proxy: ['/api'] },
} as const

type AppMode = keyof typeof apps

function selectedApp(mode: string): AppMode {
  return mode in apps ? mode as AppMode : 'diff'
}

export default defineConfig(({ mode }) => {
  const app = apps[selectedApp(mode)]
  const backend = process.env.YCY_WEB_BACKEND ?? `http://127.0.0.1:${app.port + 1000}`
  const proxy = Object.fromEntries(app.proxy.map(path => [path, { target: backend, changeOrigin: true }]))

  return {
    plugins: [react(), tailwindcss()],
    appType: 'mpa',
    base: '/',
    publicDir: false,
    build: {
      outDir: 'dist',
      emptyOutDir: true,
      manifest: true,
      assetsDir: 'assets',
      assetsInlineLimit: 0,
      rollupOptions: {
        input: {
          diff: resolve(import.meta.dirname, 'diff/index.html'),
          fs: resolve(import.meta.dirname, 'fs/index.html'),
          tunnelServer: resolve(import.meta.dirname, 'tunnel-server/index.html'),
        },
      },
    },
    server: {
      port: app.port,
      strictPort: true,
      proxy,
    },
    test: {
      environment: 'node',
      include: ['**/*.test.ts', '**/*.test.tsx'],
    },
  }
})
