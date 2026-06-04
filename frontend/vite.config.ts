import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { tanstackRouter } from '@tanstack/router-plugin/vite'

export default defineConfig({
  plugins: [
    tanstackRouter({
      target: 'react',
      generatedRouteTree: './src/routeTree.gen.ts',
      routesDirectory: './src/routes',
      semicolons: false,
      quoteStyle: 'single',
    }),
    react({
      babel: {
        plugins: ['babel-plugin-react-compiler'],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@wails': path.resolve(__dirname, './wailsjs'),
    },
  },
  server: {
    port: 34115,
    strictPort: true,
  },
})
