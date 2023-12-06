import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
    server: {
        host: '0.0.0.0',
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
            },
            '/api/v1/remote': {
                target: 'http://localhost:8080',
                ws: true
            }
        }
    },
    plugins: [react()],
})
