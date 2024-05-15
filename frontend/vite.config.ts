import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'fs'

// https://vitejs.dev/config/
export default defineConfig({
    server: {
        host: '0.0.0.0',
        https: {
            key: fs.readFileSync('../tls/server.key'),
            cert: fs.readFileSync('../tls/server.crt'),
        },
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
            },
            '/api/v1/remote': {
                target: 'http://localhost:8080',
                ws: true,
            }
        }
    },
    plugins: [react()],
})
