import { defineConfig } from 'vite'
import preact from '@preact/preset-vite'

export default defineConfig({
  plugins: [preact()],
  server: {
    // `npm run dev` meneruskan request API ke `./auth serve` yang jalan di lokal.
    proxy: {
      '/api': process.env.API_TARGET || 'http://localhost:8080',
    },
  },
})
