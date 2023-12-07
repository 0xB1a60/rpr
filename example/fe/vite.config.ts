import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import checker from 'vite-plugin-checker'
import eslint from 'vite-plugin-eslint';

// https://vitejs.dev/config/
export default defineConfig({
    server: {
        port: 9001,
        open: true,
    },
    plugins: [
        react(),
        eslint({
            failOnError: true,
            failOnWarning: false,
        }),
        checker({
            typescript: true,
            enableBuild: true,
        }),
    ],
})
