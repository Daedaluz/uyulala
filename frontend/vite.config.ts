import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'fs'

function getFile(file: string): Buffer|null {
    try {
        return fs.readFileSync(file);
    } catch (e) {
        return null
    }
}

const configs = {
    build: {
        plugins: [react()],
    },
    serve: {
        server: {
            host: '0.0.0.0',
            https: {
                key: getFile('../tls/server.key'),
                cert: getFile('../tls/server.crt'),
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
    }
}

// https://vitejs.dev/config/
export default defineConfig(({command}) => {
    return configs[command];
})
