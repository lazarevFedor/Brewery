// frontend/vite.config.js
export default {
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
  build: {
    outDir: '../static', // билд сразу в папку, которую раздаёт Gin
  }
}