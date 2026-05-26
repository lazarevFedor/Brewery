// frontend/vite.config.js

import { defineConfig } from 'vite'
import { resolve } from 'path'

export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      // Все запросы /api/* → пересылаются на Go-сервер
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      }
    }
  },

  root: '.',

  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        outDir: '../static', // билд сразу в папку, которую раздаёт Gin
      },
    },
  },

  esbuild: {
    jsxFactory: 'h',
    jsxFragment: 'Fragment',
    jsxInject: `import React from 'react'`, // если используете React
  },
})